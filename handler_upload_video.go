package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
    const uploadLimit = 10 << 30  // 1GB
    r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

    // auth user
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading video", videoID, "by user", userID)

    file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

    // verify file type 
    mimeType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get mimeType of video file to upload", err)
		return
	}

    if mimeType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Invalid Video Format", err)
		return
    }

    // get video meta data from db
    video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get video from db", err)
		return
	}

    // only video owner can modify video
    if video.UserID != userID {
        w.WriteHeader(http.StatusUnauthorized)
		return
    }

    // save file on disk
    tempVidFile, err := os.CreateTemp("","tubely-upload-video.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to save video on disk", err)
		return
	}

    _, err = io.Copy(tempVidFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to copy video file into server", err)
		return
	}

    // reset tempVidFile pointer to read again
    tempVidFile.Seek(0, io.SeekStart)

    aspRat, err  := GetVideoAspectRatio(tempVidFile.Name())
    if err != nil {
        fmt.Printf("error getting video aspect ratio: %v\n", err)
		respondWithError(w, http.StatusBadRequest, "Error Getting Video Aspect Ratio", err)
        return
    }

    // change tempvidfile to a processed fast start video
    fastStartFilePath, err := ProcessVideoForFastStart(tempVidFile.Name())
    if err != nil {
        fmt.Printf("error making a fast start processed video: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Error Making Fast Start Video", err)
        return
    }

    fastFile, err := os.OpenFile(fastStartFilePath, os.O_RDONLY, 0644)
    if err != nil {
        fmt.Printf("error opening fast start processed video: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Error Opening Fast Start Video", err)
        return
    }
    defer os.Remove(fastFile.Name())
    defer fastFile.Close()

    tempVidFile.Close()
    os.Remove(tempVidFile.Name())

    // gen random name
    rndm := make([]byte, 32)
    rand.Read(rndm)
    name := base64.RawURLEncoding.EncodeToString(rndm)

    // put video into an aws object
    key := fmt.Sprintf("%s/%s.mp4", aspRat, name)

    object := s3.PutObjectInput{
    	Bucket:         &cfg.s3Bucket,
        Key:            &key,
    	Body:           fastFile,
    	ContentType:    &mimeType,
    }

    _, err = cfg.s3Client.PutObject(
        r.Context(),
        &object,
    )

    if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to put video on aws stuff", err)
		return
    }

    url := fmt.Sprintf("%s/%s", cfg.s3CfDistribution, key)
    video.VideoURL = &url

    // update database
    err = cfg.db.UpdateVideo(video)
    if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update video on database", err)
		return
    }

	respondWithJSON(w, http.StatusOK, video)
}
