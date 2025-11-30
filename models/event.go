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

type SyncConfig struct {
	UserID string `json:"user_id" validate:"required"`
}

type RestaurantRecommendation struct {
	RestaurantName       string  `json:"restaurant_name"`
	RecommendationRating float64 `json:"recommendation_rating"`
	MainDishes           string  `json:"main_dishes"`
	ShortReason          string  `json:"short_reason"`
}

type EventAgentResponse struct {
	Recommendations []RestaurantRecommendation `json:"recommendations"`
}
