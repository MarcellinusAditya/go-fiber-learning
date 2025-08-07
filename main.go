package main

import (
	"database/sql"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type Activity struct {
	ID          int    `json:"id"`
	Title       string `json:"title" validate:"required"`
	Category    string `json:"category" validate:"required,oneof=TASK EVENT MEETING"`
	Description string `json:"description" validate:"required"`
	ActivityDate time.Time `json:"activity_date" validate:"required"`
	Status	  string `json:"status" validate:"required,oneof=NEW IN_PROGRESS COMPLETED"`
	CreatedAt   time.Time `json:"created_at"`
}

func InitDB() (*sql.DB, error) {
	
	dbs := "user=postgres.vnscudjxsqzkktnlrrho password=vitroweb host=aws-0-ap-southeast-1.pooler.supabase.com port=6543 dbname=postgres"

	db,err:= sql.Open("postgres", dbs)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	db, err := InitDB()
	if err != nil {
		panic(err)
	}	
	defer db.Close()

	app := fiber.New()

	//Get all activities
	app.Get("/activities", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, title, category, description, activity_date, status, created_at FROM activities")		
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch activities"})
		}
		defer rows.Close()

		var activities []Activity
		for rows.Next() {	
			var activity Activity
			if err := rows.Scan(&activity.ID, &activity.Title, &activity.Category, &activity.Description, &activity.ActivityDate, &activity.Status, &activity.CreatedAt); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to scan activity"})
			}
			activities = append(activities, activity)
		}
		return c.JSON(activities)
	})

	app.Post("/activities", func(c *fiber.Ctx) error {
		var activity Activity
		if err := c.BodyParser(&activity); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
			}
		
		err := validator.New().Struct(&activity)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Validation failed", "errors": err.Error()})
		}
		sqlStatement := `INSERT INTO activities (title, category, description, activity_date, status)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`
		err = db.QueryRow(sqlStatement, activity.Title, activity.Category, activity.Description, activity.ActivityDate, "NEW").Scan(&activity.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})	

		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Activity created successfully", "id": activity.ID})
	})

	// Update activity status
	app.Put("/activities/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var activity Activity
		if err := c.BodyParser(&activity); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		err := validator.New().Struct(&activity)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Validation failed", "errors": err.Error()})
		}	
		sqlStatement := `UPDATE activities SET title=$1, category=$2, description=$3, activity_date=$4, status=$5
		 WHERE id = $6 RETURNING id`

		err = db.QueryRow(sqlStatement, activity.Title, activity.Category, activity.Description, activity.ActivityDate, activity.Status, id).Scan(&activity.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update activity", "error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Activity updated successfully", "id": activity.ID})

	})

	// Delete activity
	app.Delete("/activities/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		sqlStatement := `DELETE FROM activities WHERE id = $1`
		_, err := db.Exec(sqlStatement, id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete activity", "error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Activity deleted successfully"})

	})
	app.Listen(":8001")
}