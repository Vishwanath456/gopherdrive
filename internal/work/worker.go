package worker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	pb "gopherdrive/proto"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type ProcessingJob struct {
	FileID       string
	FilePath     string
	OriginalName string // 🔥 ADD THIS
	Ctx          context.Context
}

type WorkerPool struct {
	Jobs chan ProcessingJob
	Wg   *sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	wp := &WorkerPool{
		Jobs: make(chan ProcessingJob, 100),
		Wg:   &sync.WaitGroup{},
	}

	for i := 0; i < numWorkers; i++ {
		go wp.worker(i)
	}

	log.Println("Worker pool started with", numWorkers, "workers")

	return wp
}

func (wp *WorkerPool) worker(id int) {
	log.Println("Worker started:", id)

	for job := range wp.Jobs {

		start := time.Now()

		// 🔥 Calculate file metadata
		sha, size, ext, err := calculateFileMetadata(job.FilePath, job.OriginalName)
		if err != nil {
			log.Printf("Worker %d error: %v", id, err)
			wp.Wg.Done()
			continue
		}

		// 🔥 Connect to gRPC server
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			log.Printf("gRPC connection error: %v", err)
			wp.Wg.Done()
			continue
		}

		client := pb.NewMetadataServiceClient(conn)
		_, err = client.UpdateStatus(context.Background(), &pb.UpdateRequest{
			Id:        job.FileID,
			Sha256:    sha,
			Size:      size,
			Extension: ext,
			Status:    "completed",
		})

		// 🔥 Call UpdateStatus via gRPC
		_, err = client.UpdateStatus(context.Background(), &pb.UpdateRequest{
			Id:        job.FileID,
			Sha256:    sha,
			Size:      size,
			Extension: ext,
			Status:    "completed",
		})

		if err != nil {
			log.Printf("gRPC update error: %v", err)
		}

		conn.Close()

		log.Printf("Worker %d finished file %s in %v",
			id, job.FileID, time.Since(start))

		wp.Wg.Done()
	}
}

func calculateFileMetadata(filePath string, originalName string) (string, int64, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, "", fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	hash := sha256.New()

	_, err = io.Copy(hash, file)
	if err != nil {
		return "", 0, "", fmt.Errorf("hashing failed: %w", err)
	}

	sum := hex.EncodeToString(hash.Sum(nil))

	info, err := file.Stat()
	if err != nil {
		return "", 0, "", fmt.Errorf("stat failed: %w", err)
	}

	size := info.Size()
	ext := filepath.Ext(originalName)
	return sum, size, ext, nil
}
