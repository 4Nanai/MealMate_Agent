package models

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Event struct {
	ID                    int         `json:"id"`
	UserID                string      `json:"user_id"`
	RestaurantName        string      `json:"restaurant_name"`
	Message               string      `json:"message"`
	ScheduleTime          string      `json:"schedule_time"`
	CreatedAt             string      `json:"created_at"`
	RestaurantCoordinates Coordinates `json:"restaurant_coordinates"`
}
