package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"benchmark/pkg/benconfig"
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"benchmark/ycsb"
)

type Hotel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	Country     string    `json:"country"`
	Stars       int       `json:"stars"`
	BasePrice   float64   `json:"base_price"`
	TotalRooms  int       `json:"total_rooms"`
	Lat         float64   `json:"lat"`
	Lon         float64   `json:"lon"`
	Description string    `json:"description"`
	Images      []string  `json:"images"`
	CreatedAt   time.Time `json:"created_at"`
}

type Rate struct {
	HotelID  string  `json:"hotel_id"`
	RoomType string  `json:"room_type"`
	InDate   string  `json:"in_date"`
	OutDate  string  `json:"out_date"`
	Price    float64 `json:"price"`
	Code     string  `json:"code"`
}

type Review struct {
	ID        string    `json:"id"`
	HotelID   string    `json:"hotel_id"`
	UserID    string    `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type GeoQuery struct {
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Radius   float64 `json:"radius"`
	MaxCount int     `json:"max_count"`
}

type Reservation struct {
	ID         string    `json:"id"`
	HotelID    string    `json:"hotel_id"`
	UserID     string    `json:"user_id"`
	RoomType   string    `json:"room_type"`
	CheckIn    time.Time `json:"check_in"`
	CheckOut   time.Time `json:"check_out"`
	Guests     int       `json:"guests"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"` // confirmed, cancelled, completed
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type HUser struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Preferences string    `json:"preferences"`
	CreatedAt   time.Time `json:"created_at"`
}

type RoomAvailability struct {
	HotelID     string         `json:"hotel_id"`
	Date        string         `json:"date"`       // YYYY-MM-DD
	RoomTypes   map[string]int `json:"room_types"` // room_type -> available_count
	LastUpdated time.Time      `json:"last_updated"`
}

type HUserSession struct {
	UserID      string    `json:"user_id"`
	SessionID   string    `json:"session_id"`
	LastAccess  time.Time `json:"last_access"`
	SearchPrefs string    `json:"search_preferences"`
}

type HotelWorkload struct {
	mu sync.Mutex

	Randomizer
	taskChooser *generator.Discrete
	wp          *WorkloadParameter
	// DeathStarBench-style service namespaces
	ProfileServiceNS     string // profile service - hotel profiles
	RateServiceNS        string // rate service - pricing
	GeoServiceNS         string // geo service - location queries
	RecommendServiceNS   string // recommendation service
	ReservationServiceNS string // reservation service
	ReviewServiceNS      string // review service
	UserServiceNS        string // user service
	SearchServiceNS      string // search service (in-memory cache)
	task1Count           int
	task2Count           int
	task3Count           int
	task4Count           int
	task5Count           int

	// Data for realistic simulation
	cities      []string
	countries   []string
	roomTypes   []string
	hotelNames  []string
	userNames   []string
	coordinates map[string][2]float64 // city -> [lat, lon]
}

var _ Workload = (*HotelWorkload)(nil)

func NewHotelWorkload(wp *WorkloadParameter) *HotelWorkload {
	coordinates := map[string][2]float64{
		"New York":  {40.7128, -74.0060},
		"London":    {51.5074, -0.1278},
		"Paris":     {48.8566, 2.3522},
		"Tokyo":     {35.6762, 139.6503},
		"Sydney":    {-33.8688, 151.2093},
		"Shanghai":  {31.2304, 121.4737},
		"Berlin":    {52.5200, 13.4050},
		"Dubai":     {25.2048, 55.2708},
		"Singapore": {1.3521, 103.8198},
		"Toronto":   {43.6532, -79.3832},
	}

	return &HotelWorkload{
		mu:                   sync.Mutex{},
		Randomizer:           *NewRandomizer(wp),
		taskChooser:          createTaskGenerator(wp),
		wp:                   wp,
		ProfileServiceNS:     "profile",
		RateServiceNS:        "rate",
		GeoServiceNS:         "geo",
		RecommendServiceNS:   "recommend",
		ReservationServiceNS: "reservation",
		ReviewServiceNS:      "review",
		UserServiceNS:        "user",
		SearchServiceNS:      "search",
		cities: []string{
			"New York",
			"London",
			"Paris",
			"Tokyo",
			"Sydney",
			"Shanghai",
			"Berlin",
			"Dubai",
			"Singapore",
			"Toronto",
		},
		countries: []string{
			"USA",
			"UK",
			"France",
			"Japan",
			"Australia",
			"China",
			"Germany",
			"UAE",
			"Singapore",
			"Canada",
		},
		roomTypes: []string{"Standard", "Deluxe", "Suite", "Presidential", "Executive"},
		hotelNames: []string{
			"Grand Palace",
			"Royal Inn",
			"City Center",
			"Ocean View",
			"Mountain Lodge",
			"Business Hotel",
			"Luxury Resort",
			"Budget Stay",
			"Historic Inn",
			"Modern Suites",
		},
		userNames: []string{
			"John Smith",
			"Jane Doe",
			"Mike Johnson",
			"Sarah Wilson",
			"David Brown",
			"Lisa Davis",
			"Tom Anderson",
			"Emily Garcia",
			"Chris Martinez",
			"Anna Rodriguez",
		},
		coordinates: coordinates,
	}
}

func (wl *HotelWorkload) generateHotel(hotelID string) Hotel {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	cityIdx := wl.r.Intn(len(wl.cities))
	countryIdx := wl.r.Intn(len(wl.countries))
	nameIdx := wl.r.Intn(len(wl.hotelNames))
	city := wl.cities[cityIdx]
	coord := wl.coordinates[city]

	// Add small random offset to coordinates to simulate different hotel locations
	latOffset := (wl.r.Float64() - 0.5) * 0.1 // ±0.05 degrees (~5km)
	lonOffset := (wl.r.Float64() - 0.5) * 0.1

	images := []string{
		fmt.Sprintf("https://images.hotel.com/%s/lobby.jpg", hotelID),
		fmt.Sprintf("https://images.hotel.com/%s/room.jpg", hotelID),
		fmt.Sprintf("https://images.hotel.com/%s/pool.jpg", hotelID),
	}

	return Hotel{
		ID:          hotelID,
		Name:        fmt.Sprintf("%s %s", wl.hotelNames[nameIdx], city),
		Address:     fmt.Sprintf("%d Main Street", wl.r.Intn(9999)+1),
		City:        city,
		Country:     wl.countries[countryIdx],
		Stars:       wl.r.Intn(5) + 1,
		BasePrice:   float64(wl.r.Intn(400) + 50), // $50-$450
		TotalRooms:  wl.r.Intn(200) + 50,          // 50-250 rooms
		Lat:         coord[0] + latOffset,
		Lon:         coord[1] + lonOffset,
		Description: fmt.Sprintf("A beautiful %d-star hotel in %s", wl.r.Intn(5)+1, city),
		Images:      images,
		CreatedAt:   time.Now(),
	}
}

func (wl *HotelWorkload) generateRate(
	hotelID string,
	roomType string,
	inDate, outDate string,
) Rate {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	basePrice := float64(wl.r.Intn(400) + 50)
	// Add seasonal pricing variation
	multiplier := 0.8 + wl.r.Float64()*0.6 // 0.8 to 1.4

	return Rate{
		HotelID:  hotelID,
		RoomType: roomType,
		InDate:   inDate,
		OutDate:  outDate,
		Price:    basePrice * multiplier,
		Code:     fmt.Sprintf("RATE_%d", wl.r.Intn(99999)),
	}
}

func (wl *HotelWorkload) generateReview(reviewID, hotelID, userID string) Review {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	comments := []string{
		"Excellent service and comfortable rooms",
		"Great location, close to attractions",
		"Clean and modern facilities",
		"Friendly staff and good amenities",
		"Average experience, could be better",
		"Outstanding breakfast and spa services",
	}

	return Review{
		ID:        reviewID,
		HotelID:   hotelID,
		UserID:    userID,
		Rating:    wl.r.Intn(5) + 1,
		Comment:   comments[wl.r.Intn(len(comments))],
		CreatedAt: time.Now(),
	}
}

func (wl *HotelWorkload) generateUser(userID string) HUser {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	nameIdx := wl.r.Intn(len(wl.userNames))
	name := wl.userNames[nameIdx]

	return HUser{
		ID:    userID,
		Name:  name,
		Email: fmt.Sprintf("%s@example.com", fmt.Sprintf("user%s", userID)),
		Phone: fmt.Sprintf(
			"+1-%d-%d-%d",
			wl.r.Intn(900)+100,
			wl.r.Intn(900)+100,
			wl.r.Intn(9000)+1000,
		),
		Preferences: wl.roomTypes[wl.r.Intn(len(wl.roomTypes))],
		CreatedAt:   time.Now(),
	}
}

func (wl *HotelWorkload) generateReservation(reservationID, hotelID, userID string) Reservation {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	checkIn := time.Now().AddDate(0, 0, wl.r.Intn(90)+1) // 1-90 days from now
	checkOut := checkIn.AddDate(0, 0, wl.r.Intn(10)+1)   // 1-10 days stay
	guests := wl.r.Intn(4) + 1                           // 1-4 guests
	roomType := wl.roomTypes[wl.r.Intn(len(wl.roomTypes))]
	basePrice := float64(wl.r.Intn(400) + 50)
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	totalPrice := basePrice * float64(nights) * float64(guests) * 0.9 // some discount

	return Reservation{
		ID:         reservationID,
		HotelID:    hotelID,
		UserID:     userID,
		RoomType:   roomType,
		CheckIn:    checkIn,
		CheckOut:   checkOut,
		Guests:     guests,
		TotalPrice: totalPrice,
		Status:     "confirmed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func (wl *HotelWorkload) generateRoomAvailability(hotelID string) RoomAvailability {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	roomTypes := make(map[string]int)
	for _, roomType := range wl.roomTypes {
		roomTypes[roomType] = wl.r.Intn(20) + 5 // 5-25 rooms available
	}

	return RoomAvailability{
		HotelID:     hotelID,
		Date:        time.Now().AddDate(0, 0, wl.r.Intn(30)).Format("2006-01-02"),
		RoomTypes:   roomTypes,
		LastUpdated: time.Now(),
	}
}

func (wl *HotelWorkload) generateUserSession(userID string) HUserSession {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	return HUserSession{
		UserID:     userID,
		SessionID:  fmt.Sprintf("session_%d", wl.r.Intn(999999)),
		LastAccess: time.Now(),
		SearchPrefs: fmt.Sprintf(
			"city:%s,stars:%d",
			wl.cities[wl.r.Intn(len(wl.cities))],
			wl.r.Intn(5)+1,
		),
	}
}

// GeoSearch - DeathStarBench: Search hotels by location (Geo Service)
func (wl *HotelWorkload) GeoSearch(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()

	// Generate geo query parameters
	cityIdx := wl.r.Intn(len(wl.cities))
	city := wl.cities[cityIdx]
	coord := wl.coordinates[city]

	geoQuery := GeoQuery{
		Lat:      coord[0] + (wl.r.Float64()-0.5)*0.2, // ±0.1 degree search radius
		Lon:      coord[1] + (wl.r.Float64()-0.5)*0.2,
		Radius:   float64(wl.r.Intn(20) + 5), // 5-25km radius
		MaxCount: wl.r.Intn(20) + 5,          // return 5-25 hotels
	}
	geoQueryData, _ := json.Marshal(geoQuery)

	// 1. Query Geo Service - find nearby hotels
	_, _ = db.Read(
		ctx,
		"MongoDB2",
		fmt.Sprintf("%v:nearby:%v_%v", wl.GeoServiceNS, geoQuery.Lat, geoQuery.Lon),
	)

	// 2. Query Profile Service - get hotel details for found hotels
	for i := 0; i < 3; i++ { // simulate getting 3 hotels
		hotelID := wl.NextKeyName()
		_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.ProfileServiceNS, hotelID))
	}

	// 3. Query Rate Service - get pricing for date range
	inDate := time.Now().AddDate(0, 0, wl.r.Intn(30)+1).Format("2006-01-02")
	outDate := time.Now().AddDate(0, 0, wl.r.Intn(35)+5).Format("2006-01-02")
	for i := 0; i < 3; i++ {
		hotelID := wl.NextKeyName()
		roomType := wl.roomTypes[wl.r.Intn(len(wl.roomTypes))]
		_, _ = db.Read(
			ctx,
			"Cassandra",
			fmt.Sprintf("%v:%v:%v:%v:%v", wl.RateServiceNS, hotelID, roomType, inDate, outDate),
		)
	}

	// 4. Cache search results in Search Service
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:cache:%v", wl.SearchServiceNS, userID),
		string(geoQueryData),
	)

	// 5. Update user session
	session := wl.generateUserSession(userID)
	session.SearchPrefs = fmt.Sprintf(
		"geo_search:%s_%v_%v",
		city,
		geoQuery.Radius,
		time.Now().Unix(),
	)
	sessionData, _ := json.Marshal(session)
	db.Update(ctx, "Redis", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID), string(sessionData))

	db.Commit()
}

// MakeReservation - DeathStarBench: Complete reservation flow with multiple service interactions
func (wl *HotelWorkload) MakeReservation(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	reservationID := wl.NextKeyName()
	hotelID := wl.NextKeyName()
	userID := wl.NextKeyName()

	// 1. Check user authentication (User Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID))

	// 2. Get hotel profile (Profile Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.ProfileServiceNS, hotelID))

	// 3. Check rates and availability (Rate Service)
	inDate := time.Now().AddDate(0, 0, wl.r.Intn(30)+1).Format("2006-01-02")
	outDate := time.Now().AddDate(0, 0, wl.r.Intn(35)+5).Format("2006-01-02")
	roomType := wl.roomTypes[wl.r.Intn(len(wl.roomTypes))]
	_, _ = db.Read(
		ctx,
		"Cassandra",
		fmt.Sprintf("%v:%v:%v:%v:%v", wl.RateServiceNS, hotelID, roomType, inDate, outDate),
	)

	// 4. Create reservation (Reservation Service)
	reservation := wl.generateReservation(reservationID, hotelID, userID)
	reservation.CheckIn, _ = time.Parse("2006-01-02", inDate)
	reservation.CheckOut, _ = time.Parse("2006-01-02", outDate)
	reservation.RoomType = roomType
	reservationData, _ := json.Marshal(reservation)
	db.Update(
		ctx,
		"MongoDB2",
		fmt.Sprintf("%v:%v", wl.ReservationServiceNS, reservationID),
		string(reservationData),
	)

	// 5. Update user booking history (User Service)
	userBooking := map[string]interface{}{
		"reservation_id": reservationID,
		"hotel_id":       hotelID,
		"status":         "confirmed",
		"created_at":     time.Now(),
		"total_price":    reservation.TotalPrice,
	}
	userBookingData, _ := json.Marshal(userBooking)
	db.Update(
		ctx,
		"Cassandra",
		fmt.Sprintf("%v:%v:bookings:%v", wl.UserServiceNS, userID, reservationID),
		string(userBookingData),
	)

	// 6. Update inventory/availability cache (potentially in Redis for fast lookups)
	availability := wl.generateRoomAvailability(hotelID)
	if availability.RoomTypes[reservation.RoomType] > 0 {
		availability.RoomTypes[reservation.RoomType]--
	}
	availability.LastUpdated = time.Now()
	availabilityData, _ := json.Marshal(availability)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf(
			"%v:inventory:%v:%v",
			wl.RateServiceNS,
			hotelID,
			reservation.CheckIn.Format("2006-01-02"),
		),
		string(availabilityData),
	)

	// 7. Update user session (User Service)
	session := wl.generateUserSession(userID)
	session.SearchPrefs = fmt.Sprintf("last_booking:%s", reservationID)
	sessionData, _ := json.Marshal(session)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:session:%v", wl.UserServiceNS, userID),
		string(sessionData),
	)

	db.Commit()
}

// GetRecommendation - DeathStarBench: Hotel recommendation based on user preferences and geo location
func (wl *HotelWorkload) GetRecommendation(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()

	// 1. Get user profile and preferences (User Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID))

	// 2. Get user's booking history for recommendations (User Service)
	_, _ = db.Read(ctx, "Cassandra", fmt.Sprintf("%v:%v:bookings", wl.UserServiceNS, userID))

	// 3. Query recommendation service - this would use ML algorithms in real system
	reqParam := map[string]interface{}{
		"user_id":   userID,
		"max_count": wl.r.Intn(10) + 5, // 5-15 recommendations
		"location":  wl.cities[wl.r.Intn(len(wl.cities))],
		"price_range": map[string]float64{
			"min": float64(wl.r.Intn(100) + 50),
			"max": float64(wl.r.Intn(300) + 200),
		},
	}
	reqData, _ := json.Marshal(reqParam)
	_, _ = db.Read(ctx, "Redis", fmt.Sprintf("%v:recommend:%v", wl.RecommendServiceNS, userID))

	// 4. For each recommended hotel, get profile and rates
	for i := 0; i < 5; i++ { // simulate 5 recommended hotels
		hotelID := wl.NextKeyName()

		// Get hotel profile
		_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.ProfileServiceNS, hotelID))

		// Get rates
		roomType := wl.roomTypes[wl.r.Intn(len(wl.roomTypes))]
		inDate := time.Now().AddDate(0, 0, wl.r.Intn(30)+1).Format("2006-01-02")
		outDate := time.Now().AddDate(0, 0, wl.r.Intn(35)+5).Format("2006-01-02")
		_, _ = db.Read(
			ctx,
			"Cassandra",
			fmt.Sprintf("%v:%v:%v:%v:%v", wl.RateServiceNS, hotelID, roomType, inDate, outDate),
		)

		// Get reviews for social proof
		_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v:summary", wl.ReviewServiceNS, hotelID))
	}

	// 5. Cache recommendation results
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:cache:%v", wl.RecommendServiceNS, userID),
		string(reqData),
	)

	// 6. Update user session
	session := wl.generateUserSession(userID)
	session.SearchPrefs = fmt.Sprintf("recommendation_view:%v", time.Now().Unix())
	sessionData, _ := json.Marshal(session)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:session:%v", wl.UserServiceNS, userID),
		string(sessionData),
	)

	db.Commit()
}

// ViewReservation - DeathStarBench: View reservation details
func (wl *HotelWorkload) ViewReservation(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()
	reservationID := wl.NextKeyName()

	// 1. Verify user authentication (User Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID))

	// 2. Get reservation details (Reservation Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.ReservationServiceNS, reservationID))

	// 3. Verify reservation ownership (User Service)
	_, _ = db.Read(
		ctx,
		"Cassandra",
		fmt.Sprintf("%v:%v:bookings:%v", wl.UserServiceNS, userID, reservationID),
	)

	// 4. Get hotel profile for reservation (Profile Service)
	hotelID := wl.NextKeyName()
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.ProfileServiceNS, hotelID))

	// 5. Update user session to track recent activity (User Service)
	session := wl.generateUserSession(userID)
	session.SearchPrefs = fmt.Sprintf("viewed_reservation:%s", reservationID)
	session.LastAccess = time.Now()
	sessionData, _ := json.Marshal(session)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:session:%v", wl.UserServiceNS, userID),
		string(sessionData),
	)

	db.Commit()
}

// SubmitReview - DeathStarBench: Submit hotel review and rating
func (wl *HotelWorkload) SubmitReview(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()
	hotelID := wl.NextKeyName()
	reviewID := wl.NextKeyName()

	// 1. Verify user authentication and check if user stayed at hotel (User Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID))
	_, _ = db.Read(ctx, "Cassandra", fmt.Sprintf("%v:%v:bookings", wl.UserServiceNS, userID))

	// 2. Get hotel profile to ensure it exists (Profile Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.ProfileServiceNS, hotelID))

	// 3. Submit review (Review Service)
	review := wl.generateReview(reviewID, hotelID, userID)
	reviewData, _ := json.Marshal(review)
	db.Update(
		ctx,
		"MongoDB2",
		fmt.Sprintf("%v:%v", wl.ReviewServiceNS, reviewID),
		string(reviewData),
	)

	// 4. Update hotel's review summary/aggregated rating (Review Service)
	reviewSummary := map[string]interface{}{
		"hotel_id":     hotelID,
		"avg_rating":   float64(wl.r.Intn(4)+2) + wl.r.Float64(), // 2.0 - 5.9
		"total_count":  wl.r.Intn(1000) + 10,                     // 10-1010 reviews
		"last_updated": time.Now(),
	}
	reviewSummaryData, _ := json.Marshal(reviewSummary)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:%v:summary", wl.ReviewServiceNS, hotelID),
		string(reviewSummaryData),
	)

	// 5. Update user's review history (User Service)
	userReviewRecord := map[string]interface{}{
		"review_id":  reviewID,
		"hotel_id":   hotelID,
		"rating":     review.Rating,
		"created_at": time.Now(),
	}
	userReviewData, _ := json.Marshal(userReviewRecord)
	db.Update(
		ctx,
		"Cassandra",
		fmt.Sprintf("%v:%v:reviews:%v", wl.UserServiceNS, userID, reviewID),
		string(userReviewData),
	)

	// 6. Update search cache (might affect future recommendations)
	session := wl.generateUserSession(userID)
	session.SearchPrefs = fmt.Sprintf("submitted_review:%s", reviewID)
	sessionData, _ := json.Marshal(session)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:session:%v", wl.UserServiceNS, userID),
		string(sessionData),
	)

	db.Commit()
}

func (wl *HotelWorkload) Load(ctx context.Context, opCount int,
	db ycsb.DB,
) {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}
	if opCount%benconfig.MaxLoadBatchSize != 0 {
		log.Fatalf(
			"opCount should be a multiple of MaxLoadBatchSize, opCount: %d, MaxLoadBatchSize: %d",
			opCount,
			benconfig.MaxLoadBatchSize,
		)
	}

	round := opCount / benconfig.MaxLoadBatchSize
	var aErr error
	for i := 0; i < round; i++ {
		txnDB.Start()
		for j := 0; j < benconfig.MaxLoadBatchSize; j++ {
			key := wl.NextKeyNameFromSequence()

			// Load hotel profiles (Profile Service)
			hotel := wl.generateHotel(key)
			hotelData, _ := json.Marshal(hotel)
			txnDB.Insert(
				ctx,
				"MongoDB2",
				fmt.Sprintf("%v:%v", wl.ProfileServiceNS, key),
				string(hotelData),
			)

			// Load user data (User Service)
			user := wl.generateUser(key)
			userData, _ := json.Marshal(user)
			txnDB.Insert(
				ctx,
				"MongoDB2",
				fmt.Sprintf("%v:%v", wl.UserServiceNS, key),
				string(userData),
			)

			// Load rate data for different room types (Rate Service)
			for _, roomType := range wl.roomTypes {
				inDate := time.Now().AddDate(0, 0, wl.r.Intn(30)+1).Format("2006-01-02")
				outDate := time.Now().AddDate(0, 0, wl.r.Intn(35)+5).Format("2006-01-02")
				rate := wl.generateRate(key, roomType, inDate, outDate)
				rateData, _ := json.Marshal(rate)
				txnDB.Insert(
					ctx,
					"Cassandra",
					fmt.Sprintf("%v:%v:%v:%v:%v", wl.RateServiceNS, key, roomType, inDate, outDate),
					string(rateData),
				)
			}

			// Load geo index data (Geo Service)
			geoEntry := map[string]interface{}{
				"hotel_id": key,
				"lat":      hotel.Lat,
				"lon":      hotel.Lon,
				"city":     hotel.City,
			}
			geoData, _ := json.Marshal(geoEntry)
			txnDB.Insert(
				ctx,
				"MongoDB2",
				fmt.Sprintf("%v:index:%v", wl.GeoServiceNS, key),
				string(geoData),
			)

			// Load initial room availability data (cached in Redis)
			availability := wl.generateRoomAvailability(key)
			availabilityData, _ := json.Marshal(availability)
			txnDB.Insert(
				ctx,
				"Redis",
				fmt.Sprintf("%v:inventory:%v:%v", wl.RateServiceNS, key, availability.Date),
				string(availabilityData),
			)
		}
		err := txnDB.Commit()
		if err != nil {
			aErr = err
			fmt.Printf("Error when committing transaction: %v\n", err)
		}
	}
	if aErr != nil {
		fmt.Printf("Error in Hotel Load: %v\n", aErr)
	}
}

func (wl *HotelWorkload) Run(ctx context.Context, opCount int,
	db ycsb.DB,
) {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}
	for i := 0; i < opCount; i++ {
		switch wl.NextTask() {
		case 1:
			wl.GeoSearch(ctx, txnDB)
		case 2:
			wl.GetRecommendation(ctx, txnDB)
		case 3:
			wl.MakeReservation(ctx, txnDB)
		case 4:
			wl.ViewReservation(ctx, txnDB)
		case 5:
			wl.SubmitReview(ctx, txnDB)
		default:
			panic("Invalid task")
		}
		restTime := rand.Intn(5) + 5
		time.Sleep(time.Duration(restTime) * time.Millisecond)
	}
}

func (wl *HotelWorkload) Cleanup() {}

func (wl *HotelWorkload) NeedPostCheck() bool {
	return true
}

func (wl *HotelWorkload) NeedRawDB() bool {
	return false
}

func (wl *HotelWorkload) PostCheck(context.Context, ycsb.DB, chan int) {
}

func (wl *HotelWorkload) DisplayCheckResult() {
	fmt.Printf("Task 1 (Geo Search) count: %v\n", wl.task1Count)
	fmt.Printf("Task 2 (Get Recommendation) count: %v\n", wl.task2Count)
	fmt.Printf("Task 3 (Make Reservation) count: %v\n", wl.task3Count)
	fmt.Printf("Task 4 (View Reservation) count: %v\n", wl.task4Count)
	fmt.Printf("Task 5 (Submit Review) count: %v\n", wl.task5Count)
}

func (wl *HotelWorkload) NextTask() int64 {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	idx := wl.taskChooser.Next(wl.r)
	switch idx {
	case 1:
		wl.task1Count++
	case 2:
		wl.task2Count++
	case 3:
		wl.task3Count++
	case 4:
		wl.task4Count++
	case 5:
		wl.task5Count++
	default:
		panic("Invalid task")
	}
	return idx
}

func (wl *HotelWorkload) RandomValue() string {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	value := wl.r.Intn(100000)
	return util.ToString(value)
}
