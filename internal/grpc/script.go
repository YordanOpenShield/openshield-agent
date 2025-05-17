package agentgrpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"

	"openshield-agent/proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

// Directory where scripts are stored
const scriptsDir = "scripts"

func (s *AgentServer) GetScriptChecksums(ctx context.Context, _ *emptypb.Empty) (*proto.ChecksumResponse, error) {
	files, err := os.ReadDir(scriptsDir)
	if err != nil {
		return nil, err
	}

	var checksums []*proto.ScriptChecksum
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(scriptsDir, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		checksum := sha256.Sum256(content)
		checksums = append(checksums, &proto.ScriptChecksum{
			Filename: file.Name(),
			Checksum: hex.EncodeToString(checksum[:]),
		})
	}

	return &proto.ChecksumResponse{Scripts: checksums}, nil
}

func (s *AgentServer) SendScriptFile(ctx context.Context, file *proto.FileContent) (*proto.SyncStatus, error) {
	path := filepath.Join(scriptsDir, file.Filename)
	err := os.WriteFile(path, file.Content, 0755)
	if err != nil {
		return &proto.SyncStatus{Success: false, Message: err.Error()}, nil
	}
	return &proto.SyncStatus{Success: true, Message: "File updated successfully"}, nil
}

func (s *AgentServer) DeleteScriptFile(ctx context.Context, req *proto.DeleteScriptRequest) (*proto.SyncStatus, error) {
	path := filepath.Join("scripts", filepath.Clean(req.GetFilename()))
	err := os.Remove(path)
	if err != nil {
		log.Printf("Failed to delete script %s: %v", path, err)
		return &proto.SyncStatus{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Printf("Deleted script: %s", path)
	return &proto.SyncStatus{
		Success: true,
		Message: "Script deleted successfully",
	}, nil
}
