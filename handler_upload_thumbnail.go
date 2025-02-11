package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

    mimeType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to get mimeType of thumbnail file to upload", err)
		return
	}

    if mimeType != "image/jpeg" && mimeType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid Thumbnail File", err)
		return
    }

    mediaType := strings.Split(mimeType, "/")
    imageReader := io.MultiReader(file)

    // gen random name
    rndm := make([]byte, 32)
    rand.Read(rndm)
    name := base64.RawURLEncoding.EncodeToString(rndm)

    fileName := fmt.Sprintf("%s.%s", name, mediaType[1])
    filePath := filepath.Join(cfg.assetsRoot, fileName)

    f, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create thumbnail file into server", err)
		return
	}

    _, err = io.Copy(f, imageReader)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to copy thumbnail file into server", err)
		return
	}

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

    thumbnailUrl := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, fileName)
    video.ThumbnailURL = &thumbnailUrl

    err = cfg.db.UpdateVideo(video)
    if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update video", err)
		return
    }

    updatedVid, _ := cfg.db.GetVideo(videoID)

	respondWithJSON(w, http.StatusOK, updatedVid)
}
