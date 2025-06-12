package main

import (
	"fmt"
	"math/rand"
	"time"

	"anastasia_gofman_backend/internal/entity"
)

func mockTranslatedText(base string) entity.TranslatedText {
	return entity.TranslatedText{
		EN: base + " EN",
		RU: base + " РУ",
		ES: base + " ES",
	}
}

func mockSocialLink(id int) entity.SocialLink {
	return entity.SocialLink{
		Instagram: fmt.Sprintf("https://instagram.com/user%d", id),
		Telegram:  fmt.Sprintf("https://t.me/user%d", id),
		Vkontakte: fmt.Sprintf("https://vk.com/user%d", id),
		Facebook:  fmt.Sprintf("https://facebook.com/user%d", id),
		Twitter:   fmt.Sprintf("https://twitter.com/user%d", id),
		Youtube:   fmt.Sprintf("https://youtube.com/user%d", id),
		Linkedin:  fmt.Sprintf("https://linkedin.com/in/user%d", id),
		Whatsapp:  fmt.Sprintf("https://wa.me/123456789%02d", id),
		Pinterest: fmt.Sprintf("https://pinterest.com/user%d", id),
		Behance:   fmt.Sprintf("https://behance.net/user%d", id),
	}
}

func mockContactInfo(id int) entity.ContactInfo {
	return entity.ContactInfo{
		Email: fmt.Sprintf("user%d@example.com", id),
		Phone: fmt.Sprintf("+1-555-010%d", id),
		Links: mockSocialLink(id),
	}
}

func mockAuthor(id uint) entity.Author {
	return entity.Author{
		ID:        id,
		Name:      mockTranslatedText(fmt.Sprintf("Author %d", id)),
		Bio:       mockTranslatedText(fmt.Sprintf("Biography of Author %d", id)),
		Contact:   mockContactInfo(int(id)),
		CreatedAt: time.Now().Add(-time.Duration(rand.Intn(365*24)) * time.Hour),
		UpdatedAt: time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
		Position:  int(id),
		IsActive:  rand.Intn(2) == 1,
	}
}

func mockPhoto(id uint, ownerID uint, ownerType string, isMain bool, isPreview bool, position int) entity.Photo {
	return entity.Photo{
		ID:        id,
		Path:      fmt.Sprintf("/static/photos/photo_%d_%s_%d.jpg", ownerID, ownerType, id),
		OwnerID:   ownerID,
		OwnerType: ownerType,
		Position:  position,
		IsMain:    isMain,
		IsPreview: isPreview,
		CreatedAt: time.Now().Add(-time.Duration(rand.Intn(365*24)) * time.Hour),
		UpdatedAt: time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
	}
}

func mockArt(id uint, authorID *uint, mainPhotoID, previewPhotoID *uint) entity.Art {
	return entity.Art{
		ID:             id,
		AuthorID:       authorID,
		Name:           mockTranslatedText(fmt.Sprintf("Artwork %d", id)),
		Title:          mockTranslatedText(fmt.Sprintf("Title for Artwork %d", id)),
		Description:    mockTranslatedText(fmt.Sprintf("Description for Artwork %d. Lorem ipsum dolor sit amet, consectetur adipiscing elit.", id)),
		Medium:         mockTranslatedText(fmt.Sprintf("Medium %d", rand.Intn(5)+1)),
		Technique:      mockTranslatedText(fmt.Sprintf("Technique %d", rand.Intn(3)+1)),
		DimensionX:     rand.Intn(100) + 20,  // 20-119
		DimensionY:     rand.Intn(150) + 30,  // 30-179
		Year:           2000 + rand.Intn(25), // 2000-2024
		Frame:          mockTranslatedText(fmt.Sprintf("Frame Type %d", rand.Intn(2)+1)),
		Price:          (rand.Intn(100) + 1) * 100, // 100 - 10000
		MainPhotoID:    mainPhotoID,
		PreviewPhotoID: previewPhotoID,
		Position:       int(id),
		CreatedAt:      time.Now().Add(-time.Duration(rand.Intn(365*24)) * time.Hour),
		UpdatedAt:      time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
	}
}

func mockEvent(id int, mainPhotoID, previewPhotoID *int) entity.Event {
	startDate := time.Now().Add(time.Duration(rand.Intn(30)-15) * 24 * time.Hour) // +/- 15 days from now
	endDate := startDate.Add(time.Duration(rand.Intn(7)+1) * 24 * time.Hour)      // 1-7 days duration
	return entity.Event{
		ID:             id,
		Title:          mockTranslatedText(fmt.Sprintf("Event %d", id)),
		Description:    mockTranslatedText(fmt.Sprintf("Details about Event %d. Join us for an exciting experience!", id)),
		StartDate:      startDate,
		EndDate:        endDate,
		Location:       mockTranslatedText(fmt.Sprintf("Venue %d, City %d", rand.Intn(10)+1, rand.Intn(5)+1)),
		MainPhotoID:    mainPhotoID,
		PreviewPhotoID: previewPhotoID,
		Position:       id,
		CreatedAt:      time.Now().Add(-time.Duration(rand.Intn(365*24)) * time.Hour),
		UpdatedAt:      time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	numAuthors := 5
	numArtsPerAuthor := 3
	numEvents := 4

	var authors []entity.Author
	var arts []entity.Art
	var events []entity.Event
	var photos []entity.Photo
	var photoIDCounter uint = 1

	fmt.Println("// Mock Authors")
	for i := 1; i <= numAuthors; i++ {
		author := mockAuthor(uint(i))
		authors = append(authors, author)
		fmt.Printf("%#v,\n", author)
	}

	fmt.Println("\n// Mock Arts and their Photos")
	var artIDCounter uint = 1
	for _, author := range authors {
		currentAuthorID := author.ID
		for i := 1; i <= numArtsPerAuthor; i++ {
			var mainPhotoID, previewPhotoID *uint

			// Main Photo for Art
			mainP := mockPhoto(photoIDCounter, artIDCounter, "arts", true, false, 1)
			photos = append(photos, mainP)
			mpID := mainP.ID
			mainPhotoID = &mpID
			photoIDCounter++
			fmt.Printf("%#v,\n", mainP)

			// Preview Photo for Art
			previewP := mockPhoto(photoIDCounter, artIDCounter, "arts", false, true, 1)
			photos = append(photos, previewP)
			ppID := previewP.ID
			previewPhotoID = &ppID
			photoIDCounter++
			fmt.Printf("%#v,\n", previewP)

			// Additional photos for Art
			for j := 1; j <= rand.Intn(3)+1; j++ { // 1 to 3 additional photos
				photo := mockPhoto(photoIDCounter, artIDCounter, "arts", false, false, j+1)
				photos = append(photos, photo)
				photoIDCounter++
				fmt.Printf("%#v,\n", photo)
			}

			art := mockArt(artIDCounter, &currentAuthorID, mainPhotoID, previewPhotoID)
			arts = append(arts, art)
			fmt.Printf("%#v,\n", art)
			artIDCounter++
		}
	}

	fmt.Println("\n// Mock Events and their Photos")
	for i := 1; i <= numEvents; i++ {
		var mainPhotoID, previewPhotoID *int

		// Main Photo for Event
		mainP := mockPhoto(photoIDCounter, uint(i), "events", true, false, 1)
		photos = append(photos, mainP)
		mpID := int(mainP.ID)
		mainPhotoID = &mpID
		photoIDCounter++
		fmt.Printf("%#v,\n", mainP)

		// Preview Photo for Event
		previewP := mockPhoto(photoIDCounter, uint(i), "events", false, true, 1)
		photos = append(photos, previewP)
		ppID := int(previewP.ID)
		previewPhotoID = &ppID
		photoIDCounter++
		fmt.Printf("%#v,\n", previewP)

		// Additional photos for Event
		for j := 1; j <= rand.Intn(2)+1; j++ { // 1 to 2 additional photos
			photo := mockPhoto(photoIDCounter, uint(i), "events", false, false, j+1)
			photos = append(photos, photo)
			photoIDCounter++
			fmt.Printf("%#v,\n", photo)
		}

		event := mockEvent(i, mainPhotoID, previewPhotoID)
		events = append(events, event)
		fmt.Printf("%#v,\n", event)
	}

	fmt.Println("\n// --- Summary ---")
	fmt.Printf("// Total Authors: %d\n", len(authors))
	fmt.Printf("// Total Arts: %d\n", len(arts))
	fmt.Printf("// Total Events: %d\n", len(events))
	fmt.Printf("// Total Photos: %d\n", len(photos))
}
