import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import App from './App';
import { HostServiceClient } from './proto/Host_diffServiceClientPb';
import { UploadSnapshotRequest, UploadSnapshotResponse, GetHostHistoryRequest, GetHostHistoryResponse, SnapshotInfo, CompareSnapshotsRequest, CompareSnapshotsResponse, DiffReport, OSChange } from './proto/host_diff_pb';
import { Metadata, RpcError, StatusCode } from 'grpc-web';

// Mock the gRPC client
jest.mock('./proto/Host_diffServiceClientPb', () => {
  return {
    HostServiceClient: jest.fn(() => ({
      uploadSnapshot: jest.fn(),
      getHostHistory: jest.fn(),
      compareSnapshots: jest.fn(),
    })),
  };
});

const mockUploadSnapshot = (HostServiceClient as jest.MockedClass<typeof HostServiceClient>).mock.results[0].value.uploadSnapshot;
const mockGetHostHistory = (HostServiceClient as jest.MockedClass<typeof HostServiceClient>).mock.results[0].value.getHostHistory;
const mockCompareSnapshots = (HostServiceClient as jest.MockedClass<typeof HostServiceClient>).mock.results[0].value.compareSnapshots;

describe('App Component', () => {
  beforeEach(() => {
    // Reset mocks before each test
    mockUploadSnapshot.mockReset();
    mockGetHostHistory.mockReset();
    mockCompareSnapshots.mockReset();

    // Default mock implementations
    mockUploadSnapshot.mockImplementation((req: UploadSnapshotRequest, metadata: Metadata, callback: (err: RpcError | null, response: UploadSnapshotResponse) => void) => {
      const res = new UploadSnapshotResponse();
      res.setId('1');
      res.setIpAddress('127.0.0.1');
      res.setTimestamp('2023-01-01T00:00:00Z');
      callback(null, res);
    });

    mockGetHostHistory.mockImplementation((req: GetHostHistoryRequest, metadata: Metadata, callback: (err: RpcError | null, response: GetHostHistoryResponse) => void) => {
      const res = new GetHostHistoryResponse();
      const snap1 = new SnapshotInfo();
      snap1.setId('1');
      snap1.setIpAddress('127.0.0.1');
      snap1.setTimestamp('2023-01-01T00:00:00Z');
      const snap2 = new SnapshotInfo();
      snap2.setId('2');
      snap2.setIpAddress('127.0.0.1');
      snap2.setTimestamp('2023-01-02T00:00:00Z');
      res.setSnapshotsList([snap2, snap1]); // Newest first
      callback(null, res);
    });

    mockCompareSnapshots.mockImplementation((req: CompareSnapshotsRequest, metadata: Metadata, callback: (err: RpcError | null, response: CompareSnapshotsResponse) => void) => {
      const res = new CompareSnapshotsResponse();
      const diffReport = new DiffReport();
      diffReport.setSummary('Test Diff Summary');
      const osChange = new OSChange();
      osChange.setOldname('Linux');
      osChange.setNewname('Windows');
      diffReport.setOsChanges(osChange);
      res.setReport(diffReport);
      callback(null, res);
    });
  });

  test('renders main sections', () => {
    render(<App />);
    expect(screen.getByText(/Host Diff Tool/i)).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /Upload Snapshot/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /View Host History/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /Result/i })).toBeInTheDocument();
  });

  test('handles file upload and displays success', async () => {
    render(<App />);
    const file = new File(['{}'], 'host_127.0.0.1_2023-01-01T00-00-00Z.json', { type: 'application/json' });
    const input = screen.getByLabelText(/Choose File/i);

    fireEvent.change(input, { target: { files: [file] } });

    await waitFor(() => {
      expect(mockUploadSnapshot).toHaveBeenCalledTimes(1);
      expect(screen.getByText(/Snapshot uploaded: 1/i)).toBeInTheDocument();
    });
  });

  test('fetches and displays host history', async () => {
    render(<App />);
    const ipInput = screen.getByPlaceholderText(/Enter IP Address/i);
    const getHistoryButton = screen.getByRole('button', { name: /Get History/i });

    fireEvent.change(ipInput, { target: { value: '127.0.0.1' } });
    fireEvent.click(getHistoryButton);

    await waitFor(() => {
      expect(mockGetHostHistory).toHaveBeenCalledTimes(1);
      expect(mockGetHostHistory).toHaveBeenCalledWith(expect.any(GetHostHistoryRequest), {}, expect.any(Function));
      expect(screen.getByText(/Snapshots for 127.0.0.1:/i)).toBeInTheDocument();
      expect(screen.getByText(/ID: 2, Timestamp: 2023-01-02T00:00:00Z/i)).toBeInTheDocument();
      expect(screen.getByText(/ID: 1, Timestamp: 2023-01-01T00:00:00Z/i)).toBeInTheDocument();
    });
  });

  test('selects snapshots and compares them', async () => {
    render(<App />);
    const ipInput = screen.getByPlaceholderText(/Enter IP Address/i);
    const getHistoryButton = screen.getByRole('button', { name: /Get History/i });

    fireEvent.change(ipInput, { target: { value: '127.0.0.1' } });
    fireEvent.click(getHistoryButton);

    await waitFor(() => {
      expect(screen.getByText(/ID: 2, Timestamp: 2023-01-02T00:00:00Z/i)).toBeInTheDocument();
    });

    const snapshot1 = screen.getByText(/ID: 1, Timestamp: 2023-01-01T00:00:00Z/i);
    const snapshot2 = screen.getByText(/ID: 2, Timestamp: 2023-01-02T00:00:00Z/i);

    fireEvent.click(snapshot1);
    fireEvent.click(snapshot2);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Compare Selected \(1 vs 2\)/i })).toBeInTheDocument();
    });

    const compareButton = screen.getByRole('button', { name: /Compare Selected \(1 vs 2\)/i });
    fireEvent.click(compareButton);

    await waitFor(() => {
      expect(mockCompareSnapshots).toHaveBeenCalledTimes(1);
      expect(mockCompareSnapshots).toHaveBeenCalledWith(expect.any(CompareSnapshotsRequest), {}, expect.any(Function));
      expect(screen.getByText(/Diff Report:/i)).toBeInTheDocument();
      expect(screen.getByText(/OS changed from Linux to Windows/i)).toBeInTheDocument();
    });
  });

  test('displays error on upload failure', async () => {
    mockUploadSnapshot.mockImplementationOnce((req: UploadSnapshotRequest, metadata: Metadata, callback: (err: RpcError | null, response: UploadSnapshotResponse) => void) => {
      callback(new RpcError(StatusCode.UNKNOWN, 'Upload failed', {}), new UploadSnapshotResponse());
    });

    render(<App />);
    const file = new File(['{}'], 'host_127.0.0.1_2023-01-01T00-00-00Z.json', { type: 'application/json' });
    const input = screen.getByLabelText(/Choose File/i);

    fireEvent.change(input, { target: { files: [file] } });

    await waitFor(() => {
      expect(screen.getByText(/Error: Upload failed/i)).toBeInTheDocument();
    });
  });

  test('displays error on get history failure', async () => {
    mockGetHostHistory.mockImplementationOnce((req: GetHostHistoryRequest, metadata: Metadata, callback: (err: RpcError | null, response: GetHostHistoryResponse) => void) => {
      callback(new RpcError(StatusCode.UNKNOWN, 'History fetch failed', {}), new GetHostHistoryResponse());
    });

    render(<App />);
    const ipInput = screen.getByPlaceholderText(/Enter IP Address/i);
    const getHistoryButton = screen.getByRole('button', { name: /Get History/i });

    fireEvent.change(ipInput, { target: { value: '127.0.0.1' } });
    fireEvent.click(getHistoryButton);

    await waitFor(() => {
      expect(screen.getByText(/Error: History fetch failed/i)).toBeInTheDocument();
    });
  });

  test('displays error on compare snapshots failure', async () => {
    mockCompareSnapshots.mockImplementationOnce((req: CompareSnapshotsRequest, metadata: Metadata, callback: (err: RpcError | null, response: CompareSnapshotsResponse) => void) => {
      callback(new RpcError(StatusCode.UNKNOWN, 'Comparison failed', {}), new CompareSnapshotsResponse());
    });

    render(<App />);
    const ipInput = screen.getByPlaceholderText(/Enter IP Address/i);
    const getHistoryButton = screen.getByRole('button', { name: /Get History/i });

    fireEvent.change(ipInput, { target: { value: '127.0.0.1' } });
    fireEvent.click(getHistoryButton);

    await waitFor(() => {
      expect(screen.getByText(/ID: 2, Timestamp: 2023-01-02T00:00:00Z/i)).toBeInTheDocument();
    });

    const snapshot1 = screen.getByText(/ID: 1, Timestamp: 2023-01-01T00:00:00Z/i);
    const snapshot2 = screen.getByText(/ID: 2, Timestamp: 2023-01-02T00:00:00Z/i);

    fireEvent.click(snapshot1);
    fireEvent.click(snapshot2);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Compare Selected \(1 vs 2\)/i })).toBeInTheDocument();
    });

    const compareButton = screen.getByRole('button', { name: /Compare Selected \(1 vs 2\)/i });
    fireEvent.click(compareButton);

    await waitFor(() => {
      expect(screen.getByText(/Error: Comparison failed/i)).toBeInTheDocument();
    });
  });
});