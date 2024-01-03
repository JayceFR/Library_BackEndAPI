package api

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func (s *ApiHandler) handleCreateImage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	num_of_images := r.FormValue("no")
	num, err := strconv.Atoi(num_of_images)
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	for x := 1; x <= num; x++ {
		curr_name := "image" + strconv.Itoa(x)
		file, _, err := r.FormFile(curr_name)
		type_of_file := r.FormValue("type" + strconv.Itoa(x))
		if err != nil {
			http.Error(w, "Error retrieving the file", http.StatusBadRequest)
			return err
		}
		defer file.Close()
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading the file content", http.StatusInternalServerError)
			return err
		}
		newImage := &Images{
			ID:        uuid.New(),
			Object_id: uuid.MustParse(id),
			Type:      type_of_file,
			Data:      fileBytes,
		}
		if err := s.db.Create(newImage).Error; err != nil {
			http.Error(w, "Error storing the image in the database", http.StatusInternalServerError)
			return err
		}
	}

	return s.WriteJson(w, http.StatusOK, "succes")
}

func (s *ApiHandler) handle_get_image(ctx context.Context, id string) ([]*Images, error) {
	rows, err := s.db.WithContext(ctx).
		Select("*").
		Table("images").
		Where("object_id = ?", id).
		Rows()
	if err != nil {
		return []*Images{}, err
	}
	images := []*Images{}
	for rows.Next() {
		image := Images{}
		err := rows.Scan(
			&image.ID,
			&image.Object_id,
			&image.Type,
			&image.Data,
		)
		if err != nil {
			return []*Images{}, err
		}
		images = append(images, &image)
	}
	return images, nil
}
