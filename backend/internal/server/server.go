package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/justicecaban/host-diff-tool/backend/internal/data"
	"github.com/justicecaban/host-diff-tool/backend/internal/diff"
	"github.com/justicecaban/host-diff-tool/backend/internal/validation"
	"github.com/justicecaban/host-diff-tool/proto"
)

// Server is the gRPC server.
type Server struct {
	proto.UnimplementedHostServiceServer
	db *data.DB
}

// NewServer creates a new server.
func NewServer(db *data.DB) *Server {
	return &Server{db: db}
}

// UploadSnapshot handles the UploadSnapshot RPC.
func (s *Server) UploadSnapshot(ctx context.Context, req *proto.UploadSnapshotRequest) (*proto.UploadSnapshotResponse, error) {
	parsed, err := validation.ParseFilename(req.GetFilename())
	if err != nil {
		log.Printf("UploadSnapshot error: %v", err)
		return nil, fmt.Errorf("invalid filename: %w", err)
	}

	// Validate that the file content is valid JSON
	var jsonData interface{}
	if err := json.Unmarshal(req.GetFileContent(), &jsonData); err != nil {
		log.Printf("UploadSnapshot error: %v", err)
		return nil, fmt.Errorf("invalid JSON content: %w", err)
	}

	id, err := s.db.InsertSnapshot(parsed.IPAddress, parsed.Timestamp, req.GetFileContent())
	if err != nil {
		log.Printf("UploadSnapshot error: %v", err)
		return nil, fmt.Errorf("failed to insert snapshot: %w", err)
	}

	return &proto.UploadSnapshotResponse{
		Id:        id,
		IpAddress: parsed.IPAddress,
		Timestamp: parsed.Timestamp,
	}, nil
}

// GetHostHistory handles the GetHostHistory RPC.
func (s *Server) GetHostHistory(ctx context.Context, req *proto.GetHostHistoryRequest) (*proto.GetHostHistoryResponse, error) {
	snapshots, err := s.db.GetSnapshotsByIP(req.GetIpAddress())
	if err != nil {
		log.Printf("GetHostHistory error: %v", err)
		return nil, fmt.Errorf("failed to get snapshots by IP: %w", err)
	}

	protoSnapshots := make([]*proto.SnapshotInfo, len(snapshots))
	for i, snap := range snapshots {
		protoSnapshots[i] = &proto.SnapshotInfo{
			Id:        snap.ID,
			IpAddress: snap.IPAddress,
			Timestamp: snap.Timestamp,
		}
	}

	return &proto.GetHostHistoryResponse{
		Snapshots: protoSnapshots,
	}, nil
}

// CompareSnapshots handles the CompareSnapshots RPC.
func (s *Server) CompareSnapshots(ctx context.Context, req *proto.CompareSnapshotsRequest) (*proto.CompareSnapshotsResponse, error) {
	snapA, err := s.db.GetSnapshotByID(req.GetSnapshotIdA())
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot A: %w", err)
	}
	if snapA == nil {
		return nil, fmt.Errorf("snapshot A with ID %s not found", req.GetSnapshotIdA())
	}

	snapB, err := s.db.GetSnapshotByID(req.GetSnapshotIdB())
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot B: %w", err)
	}
	if snapB == nil {
		return nil, fmt.Errorf("snapshot B with ID %s not found", req.GetSnapshotIdB())
	}

	if snapA.IPAddress != snapB.IPAddress {
		return nil, fmt.Errorf("cannot compare snapshots from different IP addresses: %s vs %s", snapA.IPAddress, snapB.IPAddress)
	}

	report, err := diff.DiffSnapshots(snapA.Data, snapB.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to diff snapshots: %w", err)
	}

	protoReport := &proto.DiffReport{
		Summary: report.Summary,
	}

	for _, s := range report.AddedServices {
		protoReport.AddedPorts = append(protoReport.AddedPorts, &proto.PortChange{
			Port:     int32(s.Port),
			Protocol: s.Protocol,
		})
	}

	for _, s := range report.RemovedServices {
		protoReport.RemovedPorts = append(protoReport.RemovedPorts, &proto.PortChange{
			Port:     int32(s.Port),
			Protocol: s.Protocol,
		})
	}

	for _, sc := range report.ChangedServices {
		protoReport.ChangedPorts = append(protoReport.ChangedPorts, &proto.PortChange{
			Port:     int32(sc.Port),
			Protocol: sc.Protocol,
			Changes:  sc.Changes,
		})
	}

	for _, cve := range report.AddedCVEs {
		protoReport.AddedCves = append(protoReport.AddedCves, &proto.CVEChange{
			CveId: cve.CVEID,
		})
	}

	for _, cve := range report.RemovedCVEs {
		protoReport.RemovedCves = append(protoReport.RemovedCves, &proto.CVEChange{
			CveId: cve.CVEID,
		})
	}

	return &proto.CompareSnapshotsResponse{
		Report: protoReport,
	}, nil
}
