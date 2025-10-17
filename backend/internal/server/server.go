package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/justicecaban/host-diff-tool/backend/internal/data"
	"github.com/justicecaban/host-diff-tool/backend/internal/diff"
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
	ipAddress, timestamp, err := parseFilename(req.GetFilename())
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

	id, err := s.db.InsertSnapshot(ipAddress, timestamp, req.GetFileContent())
	if err != nil {
		log.Printf("UploadSnapshot error: %v", err)
		return nil, fmt.Errorf("failed to insert snapshot: %w", err)
	}

	return &proto.UploadSnapshotResponse{
		Id:        id,
		IpAddress: ipAddress,
		Timestamp: timestamp,
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

// parseFilename extracts IP address and timestamp from a filename like "host_127.0.0.1_2025-10-16T12-00-00Z.json"
func parseFilename(filename string) (ipAddress, timestamp string, err error) {
	re := regexp.MustCompile(`host_([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})_([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2}Z)\.json`)
	matches := re.FindStringSubmatch(filename)

	if len(matches) != 3 {
		return "", "", fmt.Errorf("filename does not match expected format: %s", filename)
	}

	// Validate IP address octets (0-255)
	ip := matches[1]
	var octets [4]int
	if _, err := fmt.Sscanf(ip, "%d.%d.%d.%d", &octets[0], &octets[1], &octets[2], &octets[3]); err != nil {
		return "", "", fmt.Errorf("invalid IP address format: %s", ip)
	}
	for i, octet := range octets {
		if octet < 0 || octet > 255 {
			return "", "", fmt.Errorf("invalid IP address octet [%d]: %d (must be 0-255)", i, octet)
		}
	}

	// Validate timestamp date components
	timestampStr := matches[2]
	var year, month, day, hour, minute, second int
	if _, err := fmt.Sscanf(timestampStr, "%04d-%02d-%02dT%02d-%02d-%02dZ",
		&year, &month, &day, &hour, &minute, &second); err != nil {
		return "", "", fmt.Errorf("invalid timestamp format: %s", timestampStr)
	}

	// Basic validation of date/time ranges
	if month < 1 || month > 12 {
		return "", "", fmt.Errorf("invalid month: %d (must be 1-12)", month)
	}
	if day < 1 || day > 31 {
		return "", "", fmt.Errorf("invalid day: %d (must be 1-31)", day)
	}
	if hour < 0 || hour > 23 {
		return "", "", fmt.Errorf("invalid hour: %d (must be 0-23)", hour)
	}
	if minute < 0 || minute > 59 {
		return "", "", fmt.Errorf("invalid minute: %d (must be 0-59)", minute)
	}
	if second < 0 || second > 59 {
		return "", "", fmt.Errorf("invalid second: %d (must be 0-59)", second)
	}

	// Replace dashes in timestamp with colons for ISO-8601 compliance
	timestamp = timestampStr[:13] + ":" + timestampStr[14:16] + ":" + timestampStr[17:]

	return ip, timestamp, nil
}
