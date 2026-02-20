package rest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"gopherdrive/internal/repository"
	_ "gopherdrive/internal/work"
	worker "gopherdrive/internal/work"
	pb "gopherdrive/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Handler struct {
	WP   *worker.WorkerPool
	Repo repository.MetadataRepo
}

func NewHandler(wp *worker.WorkerPool, repo repository.MetadataRepo) *Handler {
	return &Handler{
		WP:   wp,
		Repo: repo,
	}
}

func (h *Handler) GetFileHandler(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Path[len("/files/"):]

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer conn.Close()

	client := pb.NewMetadataServiceClient(conn)

	resp, err := client.GetFile(r.Context(), &pb.GetRequest{
		Id: id,
	})
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	fmt.Fprintf(w, "File ID: %s\nPath: %s\nStatus: %s",
		resp.Id, resp.Filepath, resp.Status)
}
func (h *Handler) ListFilesHandler(w http.ResponseWriter, r *http.Request) {

	files, err := h.Repo.ListAll(r.Context())
	if err != nil {
		fmt.Println("ERROR from ListAll:", err)
		http.Error(w, "failed to fetch files", http.StatusInternalServerError)
		return
	}

	fmt.Println("Files fetched:", len(files)) // 🔥 ADD THIS

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	id := uuid.New().String()
	filePath := "./data/" + id

	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	writer := bufio.NewWriter(out)

	_, err = io.Copy(writer, file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	writer.Flush()

	// Step 1: Call gRPC RegisterFile
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer conn.Close()

	client := pb.NewMetadataServiceClient(conn)

	_, err = client.RegisterFile(r.Context(), &pb.RegisterRequest{
		Id:       id,
		Filename: header.Filename,
		Filepath: filePath,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Step 2: Send job to worker AFTER file written
	h.WP.Wg.Add(1)
	h.WP.Jobs <- worker.ProcessingJob{
		FileID:       id,
		FilePath:     filePath,
		OriginalName: header.Filename, // 🔥 ADD THIS
		Ctx:          r.Context(),
	}

	fmt.Fprintf(w, "File uploaded successfully.\nFile ID: %s\nOriginal Name: %s", id, header.Filename)
	slog.Info("ListFilesHandler called")
}
