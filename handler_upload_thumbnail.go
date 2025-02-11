package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

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


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

    const maxMemory = 10 << 20
    r.ParseMultipartForm(maxMemory)

    file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

    mediaType := header.Header.Get("Content-Type")
    imageReader := io.MultiReader(file)
    imageData, err := io.ReadAll(imageReader)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to read image", err)
		return
	}

    // encode image data
    encodedImgData := base64.StdEncoding.EncodeToString(imageData)
    dataURL := fmt.Sprintf("data:%s;base64,%s", mediaType, encodedImgData)

    // get video meta data from db
    video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get video from db", err)
		return
	}

    if video.UserID != userID {
        w.WriteHeader(http.StatusUnauthorized)
		return
    }

    video.ThumbnailURL = &dataURL

    err = cfg.db.UpdateVideo(video)
    if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update video", err)
		return
    }

    updatedVid, _ := cfg.db.GetVideo(videoID)

	respondWithJSON(w, http.StatusOK, updatedVid)
}
