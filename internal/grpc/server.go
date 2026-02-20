package grpc

import (
	"context"

	"gopherdrive/internal/repository"
	pb "gopherdrive/proto"
)

type Server struct {
	pb.UnimplementedMetadataServiceServer
	Repo repository.MetadataRepo
}

func NewServer(repo repository.MetadataRepo) *Server {
	return &Server{Repo: repo}
}

func (s *Server) RegisterFile(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {

	err := s.Repo.Save(ctx, &repository.Metadata{
		ID:       req.Id,
		FileName: req.Filename,
		FilePath: req.Filepath,
		Status:   "processing",
	})
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{Message: "File registered"}, nil
}

func (s *Server) UpdateStatus(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {

	err := s.Repo.Update(ctx, &repository.Metadata{
		ID:        req.Id,
		SHA256:    req.Sha256,
		Size:      req.Size,      // MUST BE HERE
		Extension: req.Extension, // MUST BE HERE
		Status:    req.Status,
	})

	if err != nil {
		return nil, err
	}

	return &pb.UpdateResponse{Message: "updated"}, nil
}
func (s *Server) GetFile(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {

	m, err := s.Repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetResponse{
		Id:       m.ID,
		Filepath: m.FilePath,
		Status:   m.Status,
	}, nil
}
