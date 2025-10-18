package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/justicecaban/host-diff-tool/proto"
)

func main() {
	// Connect to the server
	conn, err := grpc.NewClient("localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewHostServiceClient(conn)
	ctx := context.Background()

	// Test 1: Upload first snapshot
	fmt.Println("Test 1: Uploading first snapshot...")
	file1, err := os.ReadFile("./assets/host_snapshots/host_125.199.235.74_2025-09-10T03-00-00Z.json")
	if err != nil {
		log.Fatalf("Failed to read snapshot file 1: %v", err)
	}

	uploadResp1, err := client.UploadSnapshot(ctx, &pb.UploadSnapshotRequest{
		Filename:    "host_125.199.235.74_2025-09-10T03-00-00Z.json",
		FileContent: file1,
	})
	if err != nil {
		log.Fatalf("Failed to upload snapshot 1: %v", err)
	}
	fmt.Printf("✓ Snapshot 1 uploaded. ID: %s, IP: %s, Timestamp: %s\n",
		uploadResp1.Id, uploadResp1.IpAddress, uploadResp1.Timestamp)

	// Test 2: Upload second snapshot
	fmt.Println("\nTest 2: Uploading second snapshot...")
	file2, err := os.ReadFile("./assets/host_snapshots/host_125.199.235.74_2025-09-15T08-49-45Z.json")
	if err != nil {
		log.Fatalf("Failed to read snapshot file 2: %v", err)
	}

	uploadResp2, err := client.UploadSnapshot(ctx, &pb.UploadSnapshotRequest{
		Filename:    "host_125.199.235.74_2025-09-15T08-49-45Z.json",
		FileContent: file2,
	})
	if err != nil {
		log.Fatalf("Failed to upload snapshot 2: %v", err)
	}
	fmt.Printf("✓ Snapshot 2 uploaded. ID: %s, IP: %s, Timestamp: %s\n",
		uploadResp2.Id, uploadResp2.IpAddress, uploadResp2.Timestamp)

	// Test 3: Get host history
	fmt.Println("\nTest 3: Getting host history...")
	historyResp, err := client.GetHostHistory(ctx, &pb.GetHostHistoryRequest{
		IpAddress: uploadResp1.IpAddress,
	})
	if err != nil {
		log.Fatalf("Failed to get host history: %v", err)
	}
	fmt.Printf("✓ Host history retrieved. Found %d snapshots:\n", len(historyResp.Snapshots))
	for i, snap := range historyResp.Snapshots {
		fmt.Printf("  %d. ID: %s, IP: %s, Timestamp: %s\n", i+1, snap.Id, snap.IpAddress, snap.Timestamp)
	}

	// Test 4: Compare snapshots
	fmt.Println("\nTest 4: Comparing snapshots...")
	compareResp, err := client.CompareSnapshots(ctx, &pb.CompareSnapshotsRequest{
		SnapshotIdA: uploadResp1.Id,
		SnapshotIdB: uploadResp2.Id,
	})
	if err != nil {
		log.Fatalf("Failed to compare snapshots: %v", err)
	}
	fmt.Println("✓ Snapshots compared successfully")
	fmt.Println("\nDiff Report:")
	fmt.Println(compareResp.Report.Summary)

	if compareResp.Report.OsChanges != nil {
		fmt.Printf("\nOS Changes:\n  From: %s\n  To: %s\n",
			compareResp.Report.OsChanges.Oldname,
			compareResp.Report.OsChanges.Newname)
	}

	if len(compareResp.Report.AddedPorts) > 0 {
		fmt.Printf("\nAdded Ports: %d\n", len(compareResp.Report.AddedPorts))
		for _, port := range compareResp.Report.AddedPorts {
			fmt.Printf("  - Port %d (%s)\n", port.Port, port.Protocol)
		}
	}

	if len(compareResp.Report.RemovedPorts) > 0 {
		fmt.Printf("\nRemoved Ports: %d\n", len(compareResp.Report.RemovedPorts))
		for _, port := range compareResp.Report.RemovedPorts {
			fmt.Printf("  - Port %d (%s)\n", port.Port, port.Protocol)
		}
	}

	if len(compareResp.Report.ChangedPorts) > 0 {
		fmt.Printf("\nChanged Ports: %d\n", len(compareResp.Report.ChangedPorts))
		for _, port := range compareResp.Report.ChangedPorts {
			fmt.Printf("  - Port %d (%s): %v\n", port.Port, port.Protocol, port.Changes)
		}
	}

	if len(compareResp.Report.AddedCves) > 0 {
		fmt.Printf("\nAdded CVEs: %d\n", len(compareResp.Report.AddedCves))
		for _, cve := range compareResp.Report.AddedCves {
			fmt.Printf("  - %s\n", cve.CveId)
		}
	}

	if len(compareResp.Report.RemovedCves) > 0 {
		fmt.Printf("\nRemoved CVEs: %d\n", len(compareResp.Report.RemovedCves))
		for _, cve := range compareResp.Report.RemovedCves {
			fmt.Printf("  - %s\n", cve.CveId)
		}
	}

	fmt.Println("\n✅ All end-to-end tests passed!")
}
