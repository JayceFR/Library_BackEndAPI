package api

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func (s *ApiHandler) handleCreateImage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
  //retrieve the number of images from the form
	num_of_images := r.FormValue("no")
	num, err := strconv.Atoi(num_of_images)
	if err != nil {
		return err
	}
	id := r.FormValue("id")
  //Iterate through the form data fetching each image
	for x := 1; x <= num; x++ {
		curr_name := "image" + strconv.Itoa(x)
    //fetch the image 
		file, _, err := r.FormFile(curr_name)
    //fetch the type of the image example: png, jpg, bmp
		type_of_file := r.FormValue("type" + strconv.Itoa(x))
		if err != nil {
			http.Error(w, "Error retrieving the file", http.StatusBadRequest)
			return err
		}
		defer file.Close() //close the file
		fileBytes, err := io.ReadAll(file) //convert the image into bytes
		if err != nil {
			http.Error(w, "Error reading the file content", http.StatusInternalServerError)
			return err
		}
    //store the image in the database. 
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
  //Fetch the rows from the database. 
	rows, err := s.db.WithContext(ctx).
		Select("*").
		Table("images").
		Where("object_id = ?", id).
		Rows()
	if err != nil {
		return []*Images{}, err
	}
  //Create an empty array of type images
	images := []*Images{}
	for rows.Next() {
    //fetch indiviual image 
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
    //add it to the array. 
		images = append(images, &image)
	}
	return images, nil
}
