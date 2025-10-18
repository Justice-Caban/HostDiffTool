import React from 'react';
import './DiffViewer.css';

interface ServiceInfo {
  port: number;
  protocol: string;
  status?: number;
  software?: {
    vendor?: string;
    product?: string;
    version?: string;
  };
  tls?: {
    version?: string;
    cipher?: string;
    certFingerprintSha256?: string;
  };
  vulnerabilities?: string[];
}

interface ServiceChange {
  port: number;
  protocol: string;
  changesMap: Array<[string, string]>;
}

interface CVEChange {
  cveId: string;
  port: number;
  protocol: string;
}

interface DiffReport {
  addedServicesList: ServiceInfo[];
  removedServicesList: ServiceInfo[];
  changedServicesList: ServiceChange[];
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
    report.addedServicesList.length > 0 ||
    report.removedServicesList.length > 0 ||
    report.changedServicesList.length > 0 ||
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

  const formatService = (service: ServiceInfo): string => {
    let result = `Port ${service.port} (${service.protocol})`;

    if (service.status) {
      result += ` [Status: ${service.status}]`;
    }

    if (service.software?.product) {
      result += ` - ${service.software.vendor || ''}${service.software.vendor ? '/' : ''}${service.software.product}`;
      if (service.software.version) {
        result += ` v${service.software.version}`;
      }
    }

    if (service.tls) {
      result += ` [TLS: ${service.tls.version || 'unknown'}]`;
    }

    if (service.vulnerabilities && service.vulnerabilities.length > 0) {
      result += ` [${service.vulnerabilities.length} CVE${service.vulnerabilities.length > 1 ? 's' : ''}]`;
    }

    return result;
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
        {/* Removed Services */}
        {report.removedServicesList.length > 0 && (
          <div className="diff-section">
            <div className="diff-hunk-header">
              @@ Removed Services ({report.removedServicesList.length}) @@
            </div>
            {report.removedServicesList.map((service, index) => (
              <div key={`removed-${index}`} className="diff-line diff-line-removed">
                <span className="diff-line-prefix">-</span>
                <span className="diff-line-content">{formatService(service)}</span>
              </div>
            ))}
          </div>
        )}

        {/* Added Services */}
        {report.addedServicesList.length > 0 && (
          <div className="diff-section">
            <div className="diff-hunk-header">
              @@ Added Services ({report.addedServicesList.length}) @@
            </div>
            {report.addedServicesList.map((service, index) => (
              <div key={`added-${index}`} className="diff-line diff-line-added">
                <span className="diff-line-prefix">+</span>
                <span className="diff-line-content">{formatService(service)}</span>
              </div>
            ))}
          </div>
        )}

        {/* Changed Services */}
        {report.changedServicesList.length > 0 && (
          <div className="diff-section">
            <div className="diff-hunk-header">
              @@ Modified Services ({report.changedServicesList.length}) @@
            </div>
            {report.changedServicesList.map((change, index) => (
              <div key={`changed-${index}`} className="diff-change-group">
                <div className="diff-line diff-line-context">
                  <span className="diff-line-prefix"> </span>
                  <span className="diff-line-content">
                    Port {change.port} ({change.protocol}):
                  </span>
                </div>
                {change.changesMap.map(([key, value], changeIndex) => {
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
                  {cve.cveId} on Port {cve.port} ({cve.protocol})
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
                  {cve.cveId} on Port {cve.port} ({cve.protocol})
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="diff-footer">
        <div className="diff-stats">
          <span className="diff-stat-added">
            +{report.addedServicesList.length + report.addedCvesList.length} additions
          </span>
          {' '}
          <span className="diff-stat-removed">
            -{report.removedServicesList.length + report.removedCvesList.length} deletions
          </span>
          {' '}
          <span className="diff-stat-changed">
            ~{report.changedServicesList.length} modifications
          </span>
        </div>
      </div>
    </div>
  );
};

export default DiffViewer;
