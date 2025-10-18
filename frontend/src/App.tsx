import React, { useState } from 'react';
import './App.css';
import { HostServiceClient } from './proto/Host_diffServiceClientPb';
import { UploadSnapshotRequest, GetHostHistoryRequest, CompareSnapshotsRequest } from './proto/host_diff_pb';
import DiffViewer from './DiffViewer';

// Type definitions for better type safety
interface SnapshotInfo {
  id: string;
  ipAddress: string;
  timestamp: string;
}

// Match the proto PortChange.AsObject type
interface PortChange {
  port: number;
  protocol: string;
  oldState: string;
  newState: string;
  oldService: string;
  newService: string;
  changesMap: Array<[string, string]>;
}

interface CVEChange {
  cveId: string;
}

interface DiffReport {
  addedPortsList: PortChange[];
  removedPortsList: PortChange[];
  changedPortsList: PortChange[];
  addedCvesList: CVEChange[];
  removedCvesList: CVEChange[];
}

const client = new HostServiceClient('http://localhost');

function App() {
  const [ipAddress, setIpAddress] = useState<string>('');
  const [hostHistory, setHostHistory] = useState<SnapshotInfo[]>([]);
  const [selectedSnapshots, setSelectedSnapshots] = useState<string[]>([]);
  const [result, setResult] = useState<string>('');
  const [diffReport, setDiffReport] = useState<DiffReport | null>(null);
  const [showDiffViewer, setShowDiffViewer] = useState<boolean>(false);

  const handleUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target?.result as ArrayBuffer;
        const request = new UploadSnapshotRequest();
        request.setFileContent(new Uint8Array(content));
        request.setFilename(file.name);
        console.log('Uploading snapshot:', file.name);
        client.uploadSnapshot(request, {}, (err, response) => {
          if (err) {
            setResult(`Error: ${err.message}`);
          } else {
            setResult(`Snapshot uploaded: ${response.getId()}`);
            // After upload, refresh history for the IP if it's set
            if (ipAddress) {
              handleGetHistory();
            }
          }
        });
      };
      reader.readAsArrayBuffer(file);
    }
  };

  const handleGetHistory = () => {
    const request = new GetHostHistoryRequest();
    request.setIpAddress(ipAddress);
    console.log('Getting history for:', ipAddress);
    client.getHostHistory(request, {}, (err, response) => {
      if (err) {
        setResult(`Error: ${err.message}`);
        setHostHistory([]);
      } else {
        const snapshots = response.getSnapshotsList().map(s => s.toObject());
        setHostHistory(snapshots);
        setResult(''); // Clear previous result
      }
    });
  };

  const handleSelectSnapshot = (id: string) => {
    setSelectedSnapshots(prev => {
      if (prev.includes(id)) {
        return prev.filter(snapshotId => snapshotId !== id);
      } else if (prev.length < 2) {
        return [...prev, id];
      } else {
        return [prev[1], id]; // Keep only the last two selected
      }
    });
  };

  const handleCompare = () => {
    if (selectedSnapshots.length !== 2) {
      setResult('Please select exactly two snapshots to compare.');
      setShowDiffViewer(false);
      return;
    }
    const request = new CompareSnapshotsRequest();
    request.setSnapshotIdA(selectedSnapshots[0]);
    request.setSnapshotIdB(selectedSnapshots[1]);
    console.log('Comparing snapshots:', selectedSnapshots[0], selectedSnapshots[1]);
    client.compareSnapshots(request, {}, (err, response) => {
      if (err) {
        setResult(`Error: ${err.message}`);
        setShowDiffViewer(false);
        setDiffReport(null);
      } else {
        const report = response.getReport();
        if (report) {
          // Convert proto report to format expected by DiffViewer
          const formattedReport: DiffReport = {
            addedPortsList: report.getAddedPortsList().map(p => p.toObject()),
            removedPortsList: report.getRemovedPortsList().map(p => p.toObject()),
            changedPortsList: report.getChangedPortsList().map(p => p.toObject()),
            addedCvesList: report.getAddedCvesList().map(c => c.toObject()),
            removedCvesList: report.getRemovedCvesList().map(c => c.toObject())
          };

          setDiffReport(formattedReport);
          setShowDiffViewer(true);
          setResult(''); // Clear text result when showing diff viewer

          console.log('Diff report:', formattedReport);
        } else {
          setResult('No diff report received.');
          setShowDiffViewer(false);
          setDiffReport(null);
        }
      }
    });
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>Host Diff Tool</h1>

        <div className="card">
          <h2>Upload Snapshot</h2>
          <input type="file" onChange={handleUpload} />
        </div>

        <div className="card">
          <h2>View Host History</h2>
          <input
            type="text"
            placeholder="Enter IP Address"
            value={ipAddress}
            onChange={(e) => setIpAddress(e.target.value)}
          />
          <button onClick={handleGetHistory}>Get History</button>

          {hostHistory.length > 0 && (
            <div className="history-list">
              <h3>Snapshots for {ipAddress}:</h3>
              {hostHistory.map((snapshot) => (
                <div
                  key={snapshot.id}
                  className={`snapshot-item ${selectedSnapshots.includes(snapshot.id) ? 'selected' : ''}`}
                  onClick={() => handleSelectSnapshot(snapshot.id)}
                >
                  ID: {snapshot.id}, Timestamp: {snapshot.timestamp}
                </div>
              ))}
              {selectedSnapshots.length === 2 && (
                <button onClick={handleCompare}>Compare Selected ({selectedSnapshots[0]} vs {selectedSnapshots[1]})</button>
              )}
            </div>
          )}
        </div>

        <div className="card">
          <h2>Result</h2>
          {showDiffViewer && diffReport ? (
            <DiffViewer
              report={diffReport}
              snapshotIdA={selectedSnapshots[0]}
              snapshotIdB={selectedSnapshots[1]}
            />
          ) : (
            <pre>{result}</pre>
          )}
        </div>
      </header>
    </div>
  );
}

export default App;
