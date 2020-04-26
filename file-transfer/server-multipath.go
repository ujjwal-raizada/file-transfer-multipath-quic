package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	utils "./utils"
	quic "github.com/lucas-clemente/quic-go"
)

const addr = "0.0.0.0:" + utils.PORT

func main() {

	quicConfig := &quic.Config{
		CreatePaths: true,
	}

	fmt.Println("Attaching to: ", addr)
	listener, err := quic.ListenAddr(addr, utils.GenerateTLSConfig(), quicConfig)
	utils.HandleError(err)

	fmt.Println("Server started! Waiting for streams from client...")

	sess, err := listener.Accept()
	utils.HandleError(err)

	fmt.Println("session created: ", sess.RemoteAddr())

	stream, err := sess.AcceptStream()
	utils.HandleError(err)

	fmt.Println("stream created: ", stream.StreamID())

	defer stream.Close()
	fmt.Println("Connected to server, start receiving the file name and file size")
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	stream.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	fmt.Println("file size received: ", fileSize)

	stream.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	fmt.Println("file name received: ", fileName)

	newFile, err := os.Create("storage-server/" + fileName)
	utils.HandleError(err)

	defer newFile.Close()
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < utils.BUFFERSIZE {
			fmt.Println("last chunk of file.")

			_, err := io.CopyN(newFile, stream, (fileSize - receivedBytes))
			utils.HandleError(err)

			stream.Read(make([]byte, (receivedBytes + utils.BUFFERSIZE) - fileSize))
			break
		}
		_, err := io.CopyN(newFile, stream, utils.BUFFERSIZE)
		utils.HandleError(err)

		receivedBytes += utils.BUFFERSIZE

		fmt.Println("received (bytes): ", receivedBytes, "\r")
	}
	time.Sleep(2 * time.Second)
	fmt.Println("Received file completely!")
}
