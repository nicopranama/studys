package main

import (
	"api-crud/db"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Student struct {
	NIM    string `json:"nim"`
	Nama   string `json:"nama"`
	Email  string `json:"email"`
	Alamat string `json:"alamat"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func CreateStudent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var student Student
	err := json.NewDecoder(r.Body).Decode(&student)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Format data tidak valid"})
		return
	}

	// Validasi data
	if student.NIM == "" || student.Nama == "" || student.Email == "" || student.Alamat == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Semua field harus diisi"})
		return
	}

	// Dapatkan koneksi dari paket db
	database := db.Koneksi()
	defer database.Close()

	// Periksa apakah NIM sudah ada
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM students WHERE nim = ?", student.NIM).Scan(&count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Terjadi kesalahan pada database"})
		return
	}

	if count > 0 {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "NIM sudah terdaftar"})
		return
	}

	// Tambahkan siswa baru ke database
	_, err = database.Exec(
		"INSERT INTO students (nim, nama, email, alamat) VALUES (?, ?, ?, ?)",
		student.NIM, student.Nama, student.Email, student.Alamat,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Gagal menyimpan data siswa"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(student)
}

func GetAllStudents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Dapatkan koneksi dari paket db
	database := db.Koneksi()
	defer database.Close()

	rows, err := database.Query("SELECT nim, nama, email, alamat FROM students")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Gagal mengambil data siswa"})
		return
	}
	defer rows.Close()

	var students []Student

	for rows.Next() {
		var student Student
		err = rows.Scan(&student.NIM, &student.Nama, &student.Email, &student.Alamat)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Gagal membaca data siswa"})
			return
		}
		students = append(students, student)
	}

	json.NewEncoder(w).Encode(students)
}

func GetStudentByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	name := params["nama"]

	// Dapatkan koneksi dari paket db
	database := db.Koneksi()
	defer database.Close()

	// Gunakan LIKE untuk pencarian nama
	rows, err := database.Query("SELECT nim, nama, email, alamat FROM students WHERE nama LIKE ?", "%"+name+"%")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Gagal mencari data siswa"})
		return
	}
	defer rows.Close()

	var students []Student

	for rows.Next() {
		var student Student
		err = rows.Scan(&student.NIM, &student.Nama, &student.Email, &student.Alamat)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Gagal membaca data siswa"})
			return
		}
		students = append(students, student)
	}

	if len(students) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Tidak ada siswa dengan nama tersebut"})
		return
	}

	json.NewEncoder(w).Encode(students)
}

func UpdateStudent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	nim := params["nim"]

	var student Student
	err := json.NewDecoder(r.Body).Decode(&student)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Format data tidak valid"})
		return
	}

	if student.Nama == "" || student.Email == "" || student.Alamat == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Semua field harus diisi"})
		return
	}

	database := db.Koneksi()
	defer database.Close()

	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM students WHERE nim = ?", nim).Scan(&count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Terjadi kesalahan pada database"})
		return
	}

	if count == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Siswa dengan NIM tersebut tidak ditemukan"})
		return
	}

	_, err = database.Exec(
		"UPDATE students SET nama = ?, email = ?, alamat = ? WHERE nim = ?",
		student.Nama, student.Email, student.Alamat, nim,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Gagal memperbarui data siswa"})
		return
	}

	student.NIM = nim

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(student)
}

func DeleteStudent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	nim := params["nim"]

	database := db.Koneksi()
	defer database.Close()

	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM students WHERE nim = ?", nim).Scan(&count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Terjadi kesalahan pada database"})
		return
	}

	if count == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Siswa dengan NIM tersebut tidak ditemukan"})
		return
	}

	_, err = database.Exec("DELETE FROM students WHERE nim = ?", nim)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Gagal menghapus data siswa"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data siswa berhasil dihapus"})
}

func main() {
	router := mux.NewRouter()

	database := db.Koneksi()
	database.Close()

	router.HandleFunc("/api/students", CreateStudent).Methods("POST")
	router.HandleFunc("/api/students", GetAllStudents).Methods("GET")
	router.HandleFunc("/api/students/search/{nama}", GetStudentByName).Methods("GET")
	router.HandleFunc("/api/students/{nim}", UpdateStudent).Methods("PUT")
	router.HandleFunc("/api/students/{nim}", DeleteStudent).Methods("DELETE")

	fmt.Println("Server berjalan pada port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
