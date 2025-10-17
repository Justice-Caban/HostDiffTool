import * as jspb from 'google-protobuf'



export class SnapshotInfo extends jspb.Message {
  getId(): string;
  setId(value: string): SnapshotInfo;

  getIpAddress(): string;
  setIpAddress(value: string): SnapshotInfo;

  getTimestamp(): string;
  setTimestamp(value: string): SnapshotInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SnapshotInfo.AsObject;
  static toObject(includeInstance: boolean, msg: SnapshotInfo): SnapshotInfo.AsObject;
  static serializeBinaryToWriter(message: SnapshotInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SnapshotInfo;
  static deserializeBinaryFromReader(message: SnapshotInfo, reader: jspb.BinaryReader): SnapshotInfo;
}

export namespace SnapshotInfo {
  export type AsObject = {
    id: string,
    ipAddress: string,
    timestamp: string,
  }
}

export class UploadSnapshotRequest extends jspb.Message {
  getFileContent(): Uint8Array | string;
  getFileContent_asU8(): Uint8Array;
  getFileContent_asB64(): string;
  setFileContent(value: Uint8Array | string): UploadSnapshotRequest;

  getFilename(): string;
  setFilename(value: string): UploadSnapshotRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UploadSnapshotRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UploadSnapshotRequest): UploadSnapshotRequest.AsObject;
  static serializeBinaryToWriter(message: UploadSnapshotRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UploadSnapshotRequest;
  static deserializeBinaryFromReader(message: UploadSnapshotRequest, reader: jspb.BinaryReader): UploadSnapshotRequest;
}

export namespace UploadSnapshotRequest {
  export type AsObject = {
    fileContent: Uint8Array | string,
    filename: string,
  }
}

export class UploadSnapshotResponse extends jspb.Message {
  getId(): string;
  setId(value: string): UploadSnapshotResponse;

  getIpAddress(): string;
  setIpAddress(value: string): UploadSnapshotResponse;

  getTimestamp(): string;
  setTimestamp(value: string): UploadSnapshotResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UploadSnapshotResponse.AsObject;
  static toObject(includeInstance: boolean, msg: UploadSnapshotResponse): UploadSnapshotResponse.AsObject;
  static serializeBinaryToWriter(message: UploadSnapshotResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UploadSnapshotResponse;
  static deserializeBinaryFromReader(message: UploadSnapshotResponse, reader: jspb.BinaryReader): UploadSnapshotResponse;
}

export namespace UploadSnapshotResponse {
  export type AsObject = {
    id: string,
    ipAddress: string,
    timestamp: string,
  }
}

export class GetHostHistoryRequest extends jspb.Message {
  getIpAddress(): string;
  setIpAddress(value: string): GetHostHistoryRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetHostHistoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetHostHistoryRequest): GetHostHistoryRequest.AsObject;
  static serializeBinaryToWriter(message: GetHostHistoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetHostHistoryRequest;
  static deserializeBinaryFromReader(message: GetHostHistoryRequest, reader: jspb.BinaryReader): GetHostHistoryRequest;
}

export namespace GetHostHistoryRequest {
  export type AsObject = {
    ipAddress: string,
  }
}

export class GetHostHistoryResponse extends jspb.Message {
  getSnapshotsList(): Array<SnapshotInfo>;
  setSnapshotsList(value: Array<SnapshotInfo>): GetHostHistoryResponse;
  clearSnapshotsList(): GetHostHistoryResponse;
  addSnapshots(value?: SnapshotInfo, index?: number): SnapshotInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetHostHistoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetHostHistoryResponse): GetHostHistoryResponse.AsObject;
  static serializeBinaryToWriter(message: GetHostHistoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetHostHistoryResponse;
  static deserializeBinaryFromReader(message: GetHostHistoryResponse, reader: jspb.BinaryReader): GetHostHistoryResponse;
}

export namespace GetHostHistoryResponse {
  export type AsObject = {
    snapshotsList: Array<SnapshotInfo.AsObject>,
  }
}

export class CompareSnapshotsRequest extends jspb.Message {
  getSnapshotIdA(): string;
  setSnapshotIdA(value: string): CompareSnapshotsRequest;

  getSnapshotIdB(): string;
  setSnapshotIdB(value: string): CompareSnapshotsRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CompareSnapshotsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CompareSnapshotsRequest): CompareSnapshotsRequest.AsObject;
  static serializeBinaryToWriter(message: CompareSnapshotsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CompareSnapshotsRequest;
  static deserializeBinaryFromReader(message: CompareSnapshotsRequest, reader: jspb.BinaryReader): CompareSnapshotsRequest;
}

export namespace CompareSnapshotsRequest {
  export type AsObject = {
    snapshotIdA: string,
    snapshotIdB: string,
  }
}

export class DiffReport extends jspb.Message {
  getSummary(): string;
  setSummary(value: string): DiffReport;

  getOsChanges(): OSChange | undefined;
  setOsChanges(value?: OSChange): DiffReport;
  hasOsChanges(): boolean;
  clearOsChanges(): DiffReport;

  getAddedPortsList(): Array<PortChange>;
  setAddedPortsList(value: Array<PortChange>): DiffReport;
  clearAddedPortsList(): DiffReport;
  addAddedPorts(value?: PortChange, index?: number): PortChange;

  getRemovedPortsList(): Array<PortChange>;
  setRemovedPortsList(value: Array<PortChange>): DiffReport;
  clearRemovedPortsList(): DiffReport;
  addRemovedPorts(value?: PortChange, index?: number): PortChange;

  getChangedPortsList(): Array<PortChange>;
  setChangedPortsList(value: Array<PortChange>): DiffReport;
  clearChangedPortsList(): DiffReport;
  addChangedPorts(value?: PortChange, index?: number): PortChange;

  getAddedServicesList(): Array<ServiceChange>;
  setAddedServicesList(value: Array<ServiceChange>): DiffReport;
  clearAddedServicesList(): DiffReport;
  addAddedServices(value?: ServiceChange, index?: number): ServiceChange;

  getRemovedServicesList(): Array<ServiceChange>;
  setRemovedServicesList(value: Array<ServiceChange>): DiffReport;
  clearRemovedServicesList(): DiffReport;
  addRemovedServices(value?: ServiceChange, index?: number): ServiceChange;

  getChangedServicesList(): Array<ServiceChange>;
  setChangedServicesList(value: Array<ServiceChange>): DiffReport;
  clearChangedServicesList(): DiffReport;
  addChangedServices(value?: ServiceChange, index?: number): ServiceChange;

  getAddedCvesList(): Array<CVEChange>;
  setAddedCvesList(value: Array<CVEChange>): DiffReport;
  clearAddedCvesList(): DiffReport;
  addAddedCves(value?: CVEChange, index?: number): CVEChange;

  getRemovedCvesList(): Array<CVEChange>;
  setRemovedCvesList(value: Array<CVEChange>): DiffReport;
  clearRemovedCvesList(): DiffReport;
  addRemovedCves(value?: CVEChange, index?: number): CVEChange;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DiffReport.AsObject;
  static toObject(includeInstance: boolean, msg: DiffReport): DiffReport.AsObject;
  static serializeBinaryToWriter(message: DiffReport, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DiffReport;
  static deserializeBinaryFromReader(message: DiffReport, reader: jspb.BinaryReader): DiffReport;
}

export namespace DiffReport {
  export type AsObject = {
    summary: string,
    osChanges?: OSChange.AsObject,
    addedPortsList: Array<PortChange.AsObject>,
    removedPortsList: Array<PortChange.AsObject>,
    changedPortsList: Array<PortChange.AsObject>,
    addedServicesList: Array<ServiceChange.AsObject>,
    removedServicesList: Array<ServiceChange.AsObject>,
    changedServicesList: Array<ServiceChange.AsObject>,
    addedCvesList: Array<CVEChange.AsObject>,
    removedCvesList: Array<CVEChange.AsObject>,
  }
}

export class PortChange extends jspb.Message {
  getPort(): number;
  setPort(value: number): PortChange;

  getProtocol(): string;
  setProtocol(value: string): PortChange;

  getOldState(): string;
  setOldState(value: string): PortChange;

  getNewState(): string;
  setNewState(value: string): PortChange;

  getOldService(): string;
  setOldService(value: string): PortChange;

  getNewService(): string;
  setNewService(value: string): PortChange;

  getChangesMap(): jspb.Map<string, string>;
  clearChangesMap(): PortChange;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PortChange.AsObject;
  static toObject(includeInstance: boolean, msg: PortChange): PortChange.AsObject;
  static serializeBinaryToWriter(message: PortChange, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PortChange;
  static deserializeBinaryFromReader(message: PortChange, reader: jspb.BinaryReader): PortChange;
}

export namespace PortChange {
  export type AsObject = {
    port: number,
    protocol: string,
    oldState: string,
    newState: string,
    oldService: string,
    newService: string,
    changesMap: Array<[string, string]>,
  }
}

export class ServiceChange extends jspb.Message {
  getName(): string;
  setName(value: string): ServiceChange;

  getOldVersion(): string;
  setOldVersion(value: string): ServiceChange;

  getNewVersion(): string;
  setNewVersion(value: string): ServiceChange;

  getChangesMap(): jspb.Map<string, string>;
  clearChangesMap(): ServiceChange;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServiceChange.AsObject;
  static toObject(includeInstance: boolean, msg: ServiceChange): ServiceChange.AsObject;
  static serializeBinaryToWriter(message: ServiceChange, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ServiceChange;
  static deserializeBinaryFromReader(message: ServiceChange, reader: jspb.BinaryReader): ServiceChange;
}

export namespace ServiceChange {
  export type AsObject = {
    name: string,
    oldVersion: string,
    newVersion: string,
    changesMap: Array<[string, string]>,
  }
}

export class CVEChange extends jspb.Message {
  getCveId(): string;
  setCveId(value: string): CVEChange;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CVEChange.AsObject;
  static toObject(includeInstance: boolean, msg: CVEChange): CVEChange.AsObject;
  static serializeBinaryToWriter(message: CVEChange, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CVEChange;
  static deserializeBinaryFromReader(message: CVEChange, reader: jspb.BinaryReader): CVEChange;
}

export namespace CVEChange {
  export type AsObject = {
    cveId: string,
  }
}

export class OSChange extends jspb.Message {
  getOldname(): string;
  setOldname(value: string): OSChange;

  getNewname(): string;
  setNewname(value: string): OSChange;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): OSChange.AsObject;
  static toObject(includeInstance: boolean, msg: OSChange): OSChange.AsObject;
  static serializeBinaryToWriter(message: OSChange, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): OSChange;
  static deserializeBinaryFromReader(message: OSChange, reader: jspb.BinaryReader): OSChange;
}

export namespace OSChange {
  export type AsObject = {
    oldname: string,
    newname: string,
  }
}

export class CompareSnapshotsResponse extends jspb.Message {
  getReport(): DiffReport | undefined;
  setReport(value?: DiffReport): CompareSnapshotsResponse;
  hasReport(): boolean;
  clearReport(): CompareSnapshotsResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CompareSnapshotsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CompareSnapshotsResponse): CompareSnapshotsResponse.AsObject;
  static serializeBinaryToWriter(message: CompareSnapshotsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CompareSnapshotsResponse;
  static deserializeBinaryFromReader(message: CompareSnapshotsResponse, reader: jspb.BinaryReader): CompareSnapshotsResponse;
}

export namespace CompareSnapshotsResponse {
  export type AsObject = {
    report?: DiffReport.AsObject,
  }
}

