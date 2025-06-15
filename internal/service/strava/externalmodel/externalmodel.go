package externalmodel

type Athlete struct {
	Id        int64  `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Sex       string `json:"sex"`
}

type StravaWebhookEvent struct {
	ObjectType     string            `json:"object_type"`
	ObjectId       int64             `json:"object_id"`
	AspectType     string            `json:"aspect_type"`
	Updates        map[string]string `json:"updates"`
	OwnerId        int64             `json:"owner_id"`
	SubscriptionId int               `json:"subscription_id"`
	EventTime      int64             `json:"event_time"`
}

type Activity struct {
	Id                 int64   `json:"id"`
	Name               string  `json:"name"`
	Distance           float32 `json:"distance"`
	MovingTime         int     `json:"moving_time"`
	ElapsedTime        int     `json:"elapsed_time"`
	TotalElevationGain float32 `json:"total_elevation_gain"`
	SportType          string  `json:"sport_type"`
	StartDate          string  `json:"start_date"`
	Manual             bool    `json:"manual"`
	AverageSpeed       float32 `json:"average_speed"`
	MaxSpeed           float32 `json:"max_speed"`
	Description        string  `json:"description"`
}
