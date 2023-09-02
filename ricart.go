package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type File struct {
	Name    string
	IsOpen  bool
	Content string
	Mutex   sync.Mutex
}

type DistributedFileSystem struct {
	Files            map[string]*File
	FilesMutex       sync.Mutex
	Requests         []*Request
	RequestMutex     sync.Mutex
	Acknowledged     []bool
	AcknowledgeMutex sync.Mutex
	Timestamps       []int
	TimestampMutex   sync.Mutex
	LogFile          *os.File
}

type Client struct {
	ID       int
	FileName string
}

type Request struct {
	ClientID  int
	File      *File
	Timestamp int
}

func (fs *DistributedFileSystem) OpenFile(clientID int, fileName string) *File {
	fs.FilesMutex.Lock()
	defer fs.FilesMutex.Unlock()

	file, ok := fs.Files[fileName]
	if !ok {
		fileContent, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", fileName, err)
			return nil
		}

		file = &File{
			Name:    fileName,
			IsOpen:  true,
			Content: string(fileContent),
		}
		fs.Files[fileName] = file
	} else {
		file.Mutex.Lock()
		file.IsOpen = true
		file.Mutex.Unlock()
	}

	fmt.Printf("Client %d opened file %s\n", clientID, fileName)
	return file
}

func (fs *DistributedFileSystem) CloseFile(file *File) {
	file.Mutex.Lock()
	file.IsOpen = false
	file.Mutex.Unlock()
	fmt.Printf("File %s closed\n", file.Name)
}

func (fs *DistributedFileSystem) ReadFile(clientID int, file *File) {
	fs.RequestMutex.Lock()
	defer fs.RequestMutex.Unlock()

	fs.TimestampMutex.Lock()
	timestamp := len(fs.Timestamps) + 1
	fs.Timestamps = append(fs.Timestamps, timestamp)
	fs.TimestampMutex.Unlock()

	request := &Request{
		ClientID:  clientID,
		File:      file,
		Timestamp: timestamp,
	}

	fs.Requests = append(fs.Requests, request)

	for i := range fs.Requests {
		if fs.Requests[i].ClientID != clientID {
			go fs.SendRequest(clientID, fs.Requests[i])
		}
	}

	for i := range fs.Requests {
		if fs.Requests[i].ClientID != clientID {
			fs.ReceiveAcknowledge()
		}
	}

	fmt.Printf("Client %d read file %s: %s\n", clientID, file.Name, file.Content)
	fs.LogRequest(clientID, "Read", file.Name, timestamp)

}

func (fs *DistributedFileSystem) WriteFile(clientID int, file *File, content string) {
	fs.RequestMutex.Lock()
	defer fs.RequestMutex.Unlock()

	fs.TimestampMutex.Lock()
	timestamp := len(fs.Timestamps) + 1
	fs.Timestamps = append(fs.Timestamps, timestamp)
	fs.TimestampMutex.Unlock()

	request := &Request{
		ClientID:  clientID,
		File:      file,
		Timestamp: timestamp,
	}

	fs.Requests = append(fs.Requests, request)

	for i := range fs.Requests {
		if fs.Requests[i].ClientID != clientID {
			go fs.SendRequest(clientID, fs.Requests[i])
		}
	}

	for i := range fs.Requests {
		if fs.Requests[i].ClientID != clientID {
			fs.ReceiveAcknowledge()
		}
	}

	file.Mutex.Lock()
	file.Content = content
	file.Mutex.Unlock()

	err := ioutil.WriteFile(file.Name, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing to file %s: %v\n", file.Name, err)
		return
	}

	fmt.Printf("Client %d wrote to file %s: %s\n", clientID, file.Name, content)
	fs.LogRequest(clientID, "Write", file.Name, timestamp)
}

func (fs *DistributedFileSystem) SendRequest(clientID int, request *Request) {
	fmt.Printf("Client %d sent request to client %d\n", clientID, request.ClientID)
	fs.AcknowledgeMutex.Lock()
	defer fs.AcknowledgeMutex.Unlock()

	if request.ClientID < len(fs.Acknowledged) {
		fs.Acknowledged[request.ClientID] = true
	} else {
		fmt.Println("Invalid client ID in SendRequest")
	}
}

func (fs *DistributedFileSystem) ReceiveAcknowledge() {
	fs.AcknowledgeMutex.Lock()
	defer fs.AcknowledgeMutex.Unlock()

	for i := 0; i < len(fs.Acknowledged); i++ {
		if !fs.Acknowledged[i] {
			return
		}
	}

	fs.Acknowledged = make([]bool, len(fs.Requests))

	fmt.Println("All acknowledgments received")
}

func (fs *DistributedFileSystem) LogRequest(clientID int, action string, fileName string, timestamp int) {
	logEntry := fmt.Sprintf("Client %d %s file %s at timestamp %d\n", clientID, action, fileName, timestamp)
	fs.LogFile.WriteString(logEntry)
}

func main() {
	fileSystem := &DistributedFileSystem{
		Files:            make(map[string]*File),
		FilesMutex:       sync.Mutex{},
		Requests:         []*Request{},
		RequestMutex:     sync.Mutex{},
		AcknowledgeMutex: sync.Mutex{},
		Timestamps:       []int{},
		TimestampMutex:   sync.Mutex{},
	}

	logFile, err := os.OpenFile("file_access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer logFile.Close()

	fileSystem.LogFile = logFile
	var numClients int
	fmt.Print("Enter the number of clients: ")
	fmt.Scanln(&numClients)
	fileSystem.Acknowledged = make([]bool, numClients)

	var wg sync.WaitGroup

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			client := &Client{
				ID:       clientID + 1,
				FileName: "file1.txt",
			}
			file := fileSystem.OpenFile(client.ID, client.FileName)
			if file != nil {
				fileSystem.WriteFile(client.ID, file, fmt.Sprintf("Content written by Client %d", client.ID))
				fileSystem.ReadFile(client.ID, file)
				fileSystem.CloseFile(file)
			}
		}(i)
	}

	wg.Wait()
}