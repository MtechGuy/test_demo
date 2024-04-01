package postgresql

import (
	"database/sql"
	"encoding/base64"
	"net/http"

	"github.com/pkg/errors"

	models "amencia.net/ubb-campus-safety-main/pkg/model"
)

type ConnectModel struct {
	DB *sql.DB
}

func (m *ConnectModel) Insert(incidentType, personName, location, description, imageName string, imageData []byte, deviceLocation string) (int, error) {
	var id int
	s := `INSERT INTO report (type_of_incident, person_name, location, description, imagename, imagedata, device_location)
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING report_id;`

	err := m.DB.QueryRow(s, incidentType, personName, location, description, imageName, imageData, deviceLocation).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *ConnectModel) NewUser(username, fname, lastname, middlename, gender, dob string, imagedata []byte, imagename string, usertype int) (int, error) {
	// Check if username already exists
	usernameExists, err := m.usernameExists(fname + lastname)
	if err != nil {
		return 0, err
	}
	if usernameExists {
		return 0, errors.New("Username already exists")
	}

	var id int
	s := `INSERT INTO personnelinfotable(username, fname, mname, lname, dob, gender, image, imagedata)
          VALUES ($1,$2, $3, $4, $5, $6, $7, $8)
          RETURNING id;`

	err = m.DB.QueryRow(s, username, fname, middlename, lastname, dob, gender, imagename, imagedata).Scan(&id)
	if err != nil {
		return 0, err
	}

	s = `INSERT INTO login(memberid, username, password, role)
    VALUES ($1, $2, $3, $4)`

	_, err = m.DB.Exec(s, id, username, username, usertype)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// to ensure no duplicate user is being added
func (m *ConnectModel) usernameExists(username string) (bool, error) {
	var count int
	err := m.DB.QueryRow("SELECT COUNT(*) FROM login WHERE username = $1", username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ReadReport fetches reports from the database
func (m *ConnectModel) ReadReport() ([]*models.Report, error) {
	s := `
        SELECT person_name, type_of_incident, location, description, imagename, imagedata
        FROM report
    `
	rows, err := m.DB.Query(s)
	if err != nil {
		return nil, errors.Wrap(err, "error querying database")
	}
	defer rows.Close()

	reports := []*models.Report{}

	for rows.Next() {
		q := &models.Report{}
		var imagename string
		var imagedata []byte
		err := rows.Scan(&q.PersonName, &q.TypeOfIncident, &q.Location, &q.Description, &imagename, &imagedata)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning row")
		}

		q.ImageName = imagename
		q.ImageData = imagedata

		// Encode image data to base64
		q.EncodedImageData = base64.StdEncoding.EncodeToString(imagedata)

		// Determine the MIME type of the image data
		q.MimeType = http.DetectContentType(imagedata)

		reports = append(reports, q)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating over rows")
	}
	return reports, nil
}

// ConvertToBase64 converts a byte slice to a base64 encoded string
func ConvertToBase64(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil // Return empty string if data is empty
	}

	// Encode data to base64
	encodedString := base64.StdEncoding.EncodeToString(data)

	return encodedString, nil
}

func (m *ConnectModel) ReadProfile(username string) ([]*models.Profile, error) {
	// Query to fetch the member ID associated with the logged-in user
	var memberID int
	err := m.DB.QueryRow("SELECT memberID FROM LOGIN WHERE username = $1", username).Scan(&memberID)
	if err != nil {
		return nil, err
	}

	// SQL statement to fetch profile for the logged-in user based on their member ID
	s := `
    SELECT id, image, fname, mname, lname, dob, gender, imagedata
    FROM personnelinfotable
    WHERE id = $1
    `
	rows, err := m.DB.Query(s, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profile := []*models.Profile{}

	for rows.Next() {
		q := &models.Profile{}
		err = rows.Scan(&q.ID, &q.Image, &q.Fname, &q.Mname, &q.LName, &q.DOB, &q.Gender, &q.ImageData)
		if err != nil {
			return nil, errors.Wrap(err, "error Scanning row")
		}

		//encode imagedata to base64
		q.EncodedImage = base64.StdEncoding.EncodeToString(q.ImageData)

		//determine the Mime Type of the image data
		q.MimeType = http.DetectContentType(q.ImageData)

		profile = append(profile, q)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating over rows")
	}
	return profile, nil
}

func (m *ConnectModel) Insertlog(personName, logDate, logTime, checkType string) (int, error) {
	var id int
	s := `INSERT INTO log (person_name, log_date, log_time, check_type)
	VALUES ($1, $2, $3, $4) RETURNING id;`

	err := m.DB.QueryRow(s, personName, logDate, logTime, checkType).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *ConnectModel) ReadLog() ([]*models.Log, error) {
	// Query to fetch the member ID associated with the logged-in user

	// SQL statement to fetch profile for the logged-in user based on their member ID
	s := `
    SELECT person_name, log_date, log_time, check_type
    FROM log
    `
	rows, err := m.DB.Query(s)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	log := []*models.Log{}

	for rows.Next() {
		q := &models.Log{}
		err = rows.Scan(&q.PersonName, &q.LogDate, &q.LogTime, &q.CheckType)
		if err != nil {
			return nil, errors.Wrap(err, "error Scanning row")
		}

		log = append(log, q)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating over rows")
	}
	return log, nil
}
