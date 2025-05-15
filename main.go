package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	minioClient, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "secret123", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("подключение успешно")
	fmt.Println("endpoint:", minioClient.EndpointURL())

	files, err := os.ReadDir("upload")
	if err != nil {
		log.Fatalf("Ошибка чтения папки upload/: %v", err)
	}

	ctx := context.Background()
	bucketName := "music"

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".mp3" {
			path := filepath.Join("upload", file.Name())

			_, err := minioClient.FPutObject(ctx, bucketName, file.Name(), path, minio.PutObjectOptions{
				ContentType: "audio/mpeg",
			})
			if err != nil {
				log.Printf("ошибка загрузки %s: %v", file.Name(), err)
			} else {
				fmt.Printf("загружено: %s\n", file.Name())
			}
		}
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".mp3" {
			destPath := filepath.Join("download", file.Name())

			err = minioClient.FGetObject(ctx, bucketName, file.Name(), destPath, minio.GetObjectOptions{})
			if err != nil {
				log.Printf("ошибка скачивания %s: %v\n", file.Name(), err)
			} else {
				fmt.Println("скачано:", file.Name())
			}
		}
	}
}
