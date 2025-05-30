package agentgrpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"

	"openshield-agent/internal/config"
	"openshield-agent/proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AgentServer) GetConfigChecksums(ctx context.Context, _ *emptypb.Empty) (*proto.ChecksumResponse, error) {
	files, err := os.ReadDir(config.ConfigPath)
	if err != nil {
		log.Printf("[CONFIG SYNC] Failed to read configs directory: %v", err)
		return nil, err
	}

	var checksums []*proto.Checksum
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(config.ConfigPath, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("[CONFIG SYNC] Failed to read config file %s: %v", path, err)
			continue
		}
		checksum := sha256.Sum256(content)
		checksums = append(checksums, &proto.Checksum{
			Filename: file.Name(),
			Checksum: hex.EncodeToString(checksum[:]),
		})
	}

	log.Printf("[CONFIG SYNC] Returning checksums for %d configs", len(checksums))
	return &proto.ChecksumResponse{Files: checksums}, nil
}

func (s *AgentServer) SendConfigFile(ctx context.Context, file *proto.FileContent) (*proto.SyncStatus, error) {
	path := filepath.Join(config.ConfigPath, file.Filename)
	err := os.WriteFile(path, file.Content, 0755)
	if err != nil {
		log.Printf("[CONFIG SYNC] Failed to write config file %s: %v", path, err)
		return &proto.SyncStatus{Success: false, Message: err.Error()}, nil
	}
	log.Printf("[CONFIG SYNC] Config file updated: %s", path)
	return &proto.SyncStatus{Success: true, Message: "File updated successfully"}, nil
}
