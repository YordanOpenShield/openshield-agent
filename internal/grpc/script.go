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

func (s *AgentServer) GetScriptChecksums(ctx context.Context, _ *emptypb.Empty) (*proto.ChecksumResponse, error) {
	files, err := os.ReadDir(config.ScriptsPath)
	if err != nil {
		log.Printf("[SCRIPT SYNC] Failed to read scripts directory: %v", err)
		return nil, err
	}

	var checksums []*proto.Checksum
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(config.ScriptsPath, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("[SCRIPT SYNC] Failed to read script file %s: %v", path, err)
			continue
		}
		checksum := sha256.Sum256(content)
		checksums = append(checksums, &proto.Checksum{
			Filename: file.Name(),
			Checksum: hex.EncodeToString(checksum[:]),
		})
	}

	log.Printf("[SCRIPT SYNC] Returning checksums for %d scripts", len(checksums))
	return &proto.ChecksumResponse{Files: checksums}, nil
}

func (s *AgentServer) SendScriptFile(ctx context.Context, file *proto.FileContent) (*proto.SyncStatus, error) {
	path := filepath.Join(config.ScriptsPath, file.Filename)
	err := os.WriteFile(path, file.Content, 0755)
	if err != nil {
		log.Printf("[SCRIPT SYNC] Failed to write script file %s: %v", path, err)
		return &proto.SyncStatus{Success: false, Message: err.Error()}, nil
	}
	log.Printf("[SCRIPT SYNC] Script file updated: %s", path)
	return &proto.SyncStatus{Success: true, Message: "File updated successfully"}, nil
}

func (s *AgentServer) DeleteScriptFile(ctx context.Context, req *proto.DeleteScriptRequest) (*proto.SyncStatus, error) {
	path := filepath.Join("scripts", filepath.Clean(req.GetFilename()))
	err := os.Remove(path)
	if err != nil {
		log.Printf("[SCRIPT SYNC] Failed to delete script %s: %v", path, err)
		return &proto.SyncStatus{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Printf("[SCRIPT SYNC] Deleted script: %s", path)
	return &proto.SyncStatus{
		Success: true,
		Message: "Script deleted successfully",
	}, nil
}
