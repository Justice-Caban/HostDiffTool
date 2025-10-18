#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Import the generated gRPC-Web client code
const { HostServiceClient } = require('./frontend/src/proto/Host_diffServiceClientPb');
const {
  UploadSnapshotRequest,
  GetHostHistoryRequest,
  CompareSnapshotsRequest
} = require('./frontend/src/proto/host_diff_pb');

// Create a client instance
const client = new HostServiceClient('http://localhost:8080', null, null);

async function runTests() {
  try {
    console.log('Starting E2E tests...\n');

    // Test 1: Upload first snapshot
    console.log('Test 1: Uploading first snapshot...');
    const file1Path = './assets/host_snapshots/host_125.199.235.74_2025-09-10T03-00-00Z.json';
    const file1Content = fs.readFileSync(file1Path);
    const filename1 = path.basename(file1Path);

    const uploadReq1 = new UploadSnapshotRequest();
    uploadReq1.setFilename(filename1);
    uploadReq1.setFileContent(file1Content);

    const uploadResp1 = await new Promise((resolve, reject) => {
      client.uploadSnapshot(uploadReq1, {}, (err, response) => {
        if (err) reject(err);
        else resolve(response);
      });
    });

    console.log(`✓ Snapshot 1 uploaded. ID: ${uploadResp1.getId()}, IP: ${uploadResp1.getIpAddress()}, Timestamp: ${uploadResp1.getTimestamp()}`);
    const snapshotId1 = uploadResp1.getId();

    // Test 2: Upload second snapshot
    console.log('\nTest 2: Uploading second snapshot...');
    const file2Path = './assets/host_snapshots/host_125.199.235.74_2025-09-15T08-49-45Z.json';
    const file2Content = fs.readFileSync(file2Path);
    const filename2 = path.basename(file2Path);

    const uploadReq2 = new UploadSnapshotRequest();
    uploadReq2.setFilename(filename2);
    uploadReq2.setFileContent(file2Content);

    const uploadResp2 = await new Promise((resolve, reject) => {
      client.uploadSnapshot(uploadReq2, {}, (err, response) => {
        if (err) reject(err);
        else resolve(response);
      });
    });

    console.log(`✓ Snapshot 2 uploaded. ID: ${uploadResp2.getId()}, IP: ${uploadResp2.getIpAddress()}, Timestamp: ${uploadResp2.getTimestamp()}`);
    const snapshotId2 = uploadResp2.getId();
    const ipAddress = uploadResp2.getIpAddress();

    // Test 3: Get host history
    console.log('\nTest 3: Getting host history...');
    const historyReq = new GetHostHistoryRequest();
    historyReq.setIpAddress(ipAddress);

    const historyResp = await new Promise((resolve, reject) => {
      client.getHostHistory(historyReq, {}, (err, response) => {
        if (err) reject(err);
        else resolve(response);
      });
    });

    const snapshots = historyResp.getSnapshotsList();
    console.log(`✓ Host history retrieved. Found ${snapshots.length} snapshots:`);
    snapshots.forEach((snap, i) => {
      console.log(`  ${i + 1}. ID: ${snap.getId()}, IP: ${snap.getIpAddress()}, Timestamp: ${snap.getTimestamp()}`);
    });

    // Test 4: Compare snapshots
    console.log('\nTest 4: Comparing snapshots...');
    const compareReq = new CompareSnapshotsRequest();
    compareReq.setSnapshotIdA(snapshotId1);
    compareReq.setSnapshotIdB(snapshotId2);

    const compareResp = await new Promise((resolve, reject) => {
      client.compareSnapshots(compareReq, {}, (err, response) => {
        if (err) reject(err);
        else resolve(response);
      });
    });

    console.log('✓ Snapshots compared successfully');
    const report = compareResp.getReport();
    console.log('\nDiff Report:');
    console.log(report.getSummary());

    if (report.getOsChanges()) {
      const osChanges = report.getOsChanges();
      console.log(`\nOS Changes:`);
      console.log(`  From: ${osChanges.getOldname()}`);
      console.log(`  To: ${osChanges.getNewname()}`);
    }

    const addedPorts = report.getAddedPortsList();
    if (addedPorts.length > 0) {
      console.log(`\nAdded Ports: ${addedPorts.length}`);
      addedPorts.forEach(port => {
        console.log(`  - Port ${port.getPort()} (${port.getProtocol()})`);
      });
    }

    const removedPorts = report.getRemovedPortsList();
    if (removedPorts.length > 0) {
      console.log(`\nRemoved Ports: ${removedPorts.length}`);
      removedPorts.forEach(port => {
        console.log(`  - Port ${port.getPort()} (${port.getProtocol()})`);
      });
    }

    const changedPorts = report.getChangedPortsList();
    if (changedPorts.length > 0) {
      console.log(`\nChanged Ports: ${changedPorts.length}`);
      changedPorts.forEach(port => {
        const changes = port.getChangesMap();
        console.log(`  - Port ${port.getPort()} (${port.getProtocol()}): ${JSON.stringify(Object.fromEntries(changes.entries()))}`);
      });
    }

    const addedCves = report.getAddedCvesList();
    if (addedCves.length > 0) {
      console.log(`\nAdded CVEs: ${addedCves.length}`);
      addedCves.forEach(cve => {
        console.log(`  - ${cve.getCveId()}`);
      });
    }

    const removedCves = report.getRemovedCvesList();
    if (removedCves.length > 0) {
      console.log(`\nRemoved CVEs: ${removedCves.length}`);
      removedCves.forEach(cve => {
        console.log(`  - ${cve.getCveId()}`);
      });
    }

    console.log('\n✅ All end-to-end tests passed!');
    process.exit(0);
  } catch (error) {
    console.error('\n❌ Test failed:', error.message);
    console.error(error);
    process.exit(1);
  }
}

runTests();
