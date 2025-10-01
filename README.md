# Distributed File System using Ricart-Agarwala Algorithm (Go)

A distributed file system simulator in Go that implements the Ricart-Agarwala mutual exclusion algorithm for coordinating access to shared files among multiple clients.

## Features

- Implements Ricart-Agarwala mutual exclusion for distributed systems
- Concurrent read/write access to files by multiple clients
- Timestamp-based request ordering
- Acknowledgment system for request coordination
- Logging of all file operations in file_access.log

## How It Works

- Each client that wants to access a file sends a timestamped request to all other clients.
- Clients acknowledge requests if they are not in their critical section or have lower priority.
- A client enters the critical section (reads/writes the file) only after receiving acknowledgments from all other clients.
- Operations are logged with timestamps for traceability.
- This simulates a distributed mutual exclusion system, ensuring that only one client writes to the file at a time, while coordinating reads/writes safely.

## Components

`File`: Represents a file with content and a mutex lock.

`DistributedFileSystem`: Manages files, requests, timestamps, acknowledgments, and logs.

`Client`: Represents a client accessing the system.

`Request`: Represents a timestamped request for accessing a file.

## Usage

Build the project

```go
go build main.go
```

Run the program

```bash
./main
```

Enter the number of clients when prompted.<br/>
Observe the logs in `file_access.log` to see the order of access and acknowledgments.

Example Output:
```
Enter the number of clients: 3

Client 1 opened file file1.txt

Client 2 opened file file1.txt

Client 3 opened file file1.txt

Client 1 sent request to client 2

Client 1 sent request to client 3

Client 2 sent request to client 1

Client 3 sent request to client 1

Client 1 wrote to file file1.txt: Content written by Client 1

Client 1 read file file1.txt: Content written by Client 1

File file1.txt closed
```

### Dependencies

Go 1.20+ (or latest stable version)<br/>
Standard Go libraries (sync, os, fmt, io/ioutil)

### Notes

This is a simulated distributed system, all clients run locally.<br/>
The Ricart-Agarwala algorithm ensures mutual exclusion using timestamps and acknowledgments.<br/>
The log file helps visualize request ordering and file access coordination.<br/>
