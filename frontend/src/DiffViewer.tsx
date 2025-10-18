import React from 'react';
import './DiffViewer.css';

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

interface DiffViewerProps {
  report: DiffReport | null;
  snapshotIdA: string;
  snapshotIdB: string;
}

const DiffViewer: React.FC<DiffViewerProps> = ({ report, snapshotIdA, snapshotIdB }) => {
  if (!report) {
    return (
      <div className="diff-viewer">
        <div className="diff-empty">No comparison data available</div>
      </div>
    );
  }

  const hasChanges =
    report.addedPortsList.length > 0 ||
    report.removedPortsList.length > 0 ||
    report.changedPortsList.length > 0 ||
    report.addedCvesList.length > 0 ||
    report.removedCvesList.length > 0;

  if (!hasChanges) {
    return (
      <div className="diff-viewer">
        <div className="diff-header">
          <div className="diff-file-header">
            diff --snapshot {snapshotIdA} {snapshotIdB}
          </div>
          <div className="diff-index">
            index {snapshotIdA}..{snapshotIdB}
          </div>
        </div>
        <div className="diff-empty">No meaningful differences found</div>
      </div>
    );
  }

  const formatPort = (portChange: PortChange): string => {
    return `Port ${portChange.port} (${portChange.protocol})`;
  };

  return (
    <div className="diff-viewer">
      <div className="diff-header">
        <div className="diff-file-header">
          diff --snapshot {snapshotIdA} {snapshotIdB}
        </div>
        <div className="diff-index">
          index {snapshotIdA}..{snapshotIdB}
        </div>
        <div className="diff-file-path">
          --- a/snapshot_{snapshotIdA}
        </div>
        <div className="diff-file-path">
          +++ b/snapshot_{snapshotIdB}
        </div>
      </div>

      <div className="diff-content">
        {/* Removed Ports */}
        {report.removedPortsList.length > 0 && (
          <div className="diff-section">
            <div className="diff-hunk-header">
              @@ Removed Ports ({report.removedPortsList.length}) @@
            </div>
            {report.removedPortsList.map((port, index) => (
              <div key={`removed-${index}`} className="diff-line diff-line-removed">
                <span className="diff-line-prefix">-</span>
                <span className="diff-line-content">{formatPort(port)}</span>
              </div>
            ))}
          </div>
        )}

        {/* Added Ports */}
        {report.addedPortsList.length > 0 && (
          <div className="diff-section">
            <div className="diff-hunk-header">
              @@ Added Ports ({report.addedPortsList.length}) @@
            </div>
            {report.addedPortsList.map((port, index) => (
              <div key={`added-${index}`} className="diff-line diff-line-added">
                <span className="diff-line-prefix">+</span>
                <span className="diff-line-content">{formatPort(port)}</span>
              </div>
            ))}
          </div>
        )}

        {/* Changed Ports */}
        {report.changedPortsList.length > 0 && (
          <div className="diff-section">
            <div className="diff-hunk-header">
              @@ Modified Ports ({report.changedPortsList.length}) @@
            </div>
            {report.changedPortsList.map((change, index) => (
              <div key={`changed-${index}`} className="diff-change-group">
                <div className="diff-line diff-line-context">
                  <span className="diff-line-prefix"> </span>
                  <span className="diff-line-content">
                    Port {change.port} ({change.protocol}):
                  </span>
                </div>
                {change.changesMap && change.changesMap.map(([key, value], changeIndex) => {
                  const [oldValue, newValue] = value.includes(' -> ')
                    ? value.split(' -> ')
                    : ['', value];

                  return (
                    <div key={`change-${index}-${changeIndex}`} className="diff-change-detail">
                      {oldValue && (
                        <div className="diff-line diff-line-removed">
                          <span className="diff-line-prefix">-</span>
                          <span className="diff-line-content">
                            {key}: {oldValue}
                          </span>
                        </div>
                      )}
                      <div className="diff-line diff-line-added">
                        <span className="diff-line-prefix">+</span>
                        <span className="diff-line-content">
                          {key}: {newValue}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            ))}
          </div>
        )}

        {/* Removed CVEs */}
        {report.removedCvesList.length > 0 && (
          <div className="diff-section">
            <div className="diff-hunk-header">
              @@ Removed Vulnerabilities ({report.removedCvesList.length}) @@
            </div>
            {report.removedCvesList.map((cve, index) => (
              <div key={`removed-cve-${index}`} className="diff-line diff-line-removed">
                <span className="diff-line-prefix">-</span>
                <span className="diff-line-content">
                  {cve.cveId}
                </span>
              </div>
            ))}
          </div>
        )}

        {/* Added CVEs */}
        {report.addedCvesList.length > 0 && (
          <div className="diff-section">
            <div className="diff-hunk-header">
              @@ Added Vulnerabilities ({report.addedCvesList.length}) @@
            </div>
            {report.addedCvesList.map((cve, index) => (
              <div key={`added-cve-${index}`} className="diff-line diff-line-added">
                <span className="diff-line-prefix">+</span>
                <span className="diff-line-content">
                  {cve.cveId}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="diff-footer">
        <div className="diff-stats">
          <span className="diff-stat-added">
            +{report.addedPortsList.length + report.addedCvesList.length} additions
          </span>
          {' '}
          <span className="diff-stat-removed">
            -{report.removedPortsList.length + report.removedCvesList.length} deletions
          </span>
          {' '}
          <span className="diff-stat-changed">
            ~{report.changedPortsList.length} modifications
          </span>
        </div>
      </div>
    </div>
  );
};

export default DiffViewer;
