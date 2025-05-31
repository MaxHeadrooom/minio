package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	bucketName  = "music"
	objectName  = "Fast Lane Drift.mp3"
	minDuration = 10 * time.Second
)

func main() {
	minioClient, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "secret123", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln("❌ Ошибка подключения к MinIO:", err)
	}

	log.Println("✅ Подключение к MinIO успешно")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		expiry := 20 * time.Second

		log.Println("📦 Проверка объекта:", objectName)
		info, err := minioClient.StatObject(context.Background(), bucketName, objectName, minio.StatObjectOptions{})
		if err != nil {
			log.Printf("❌ Объект не найден или ошибка при StatObject: %v", err)
			http.Error(w, "Ошибка: объект не найден или повреждён", 404)
			return
		}
		log.Printf("ℹ️ Объект найден: размер = %d байт, дата = %v", info.Size, info.LastModified)

		presignedURL, err := getPresignedURL(minioClient, bucketName, objectName, expiry)
		if err != nil {
			log.Printf("❌ Ошибка генерации pre-signed ссылки: %v", err)
			http.Error(w, "Ошибка генерации ссылки: "+err.Error(), 500)
			return
		}

		log.Printf("🔗 Ссылка сгенерирована (экспирация %v секунд): %s", expiry.Seconds(), presignedURL)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html lang="ru">
			<head><meta charset="UTF-8"><title>Прослушать трек</title></head>
			<body>
				<h1>Воспроизведение: %s</h1>
				<audio controls autoplay>
					<source src="%s" type="audio/mpeg">
					Ваш браузер не поддерживает аудио.
				</audio>
			</body>
			</html>
		`, objectName, presignedURL)
	})

	log.Println("🚀 Сервер запущен на :8080")
	http.ListenAndServe(":8080", nil)
}

func getPresignedURL(client *minio.Client, bucket, object string, expiry time.Duration) (string, error) {
	ctx := context.Background()
	reqParams := make(url.Values)

	if expiry < minDuration {
		expiry = minDuration
	}

	presignedURL, err := client.PresignedGetObject(ctx, bucket, object, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
