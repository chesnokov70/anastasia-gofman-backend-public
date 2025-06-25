package service

import (
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/repository"
)

type artCollectionService struct {
	artCollectionRepository repository.ArtCollectionRepository
}

func NewArtCollectionService(artCollectionRepository repository.ArtCollectionRepository) ArtCollectionService {
	return &artCollectionService{artCollectionRepository: artCollectionRepository}
}

func (s *artCollectionService) GetAllCollections(sorting string, with_arts bool) ([]entity.ArtCollection, error) {
	collections, err := s.artCollectionRepository.GetAllCollections(sorting, with_arts)
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (s *artCollectionService) GetCollectionByID(id uint, with_arts bool) (entity.ArtCollection, error) {
	collection, err := s.artCollectionRepository.GetCollectionByID(id, with_arts)
	if err != nil {
		return entity.ArtCollection{}, err
	}
	return collection, nil
}

func (s *artCollectionService) CreateCollection(collection entity.ArtCollection) (entity.ArtCollection, error) {
	collection, err := s.artCollectionRepository.CreateCollection(collection)
	if err != nil {
		return entity.ArtCollection{}, err
	}
	return collection, nil
}

func (s *artCollectionService) UpdateCollection(collection entity.ArtCollection) (entity.ArtCollection, error) {
	collection, err := s.artCollectionRepository.UpdateCollection(collection)
	if err != nil {
		return entity.ArtCollection{}, err
	}
	return collection, nil
}

func (s *artCollectionService) DeleteCollection(id uint, deleteAction string, artService ArtService) error {
	if deleteAction == "DELETE_ART" {

		if err := artService.DeleteArtsByCollectionIDSync(id); err != nil {
			return err
		}
	}

	return s.artCollectionRepository.DeleteCollection(id, deleteAction)
}

func (s *artCollectionService) PartialUpdateCollection(id uint, kwargs map[string]interface{}) (entity.ArtCollection, error) {
	collection, err := s.artCollectionRepository.PartialUpdateCollection(id, kwargs)
	if err != nil {
		return entity.ArtCollection{}, err
	}
	return collection, nil
}

func (s *artCollectionService) FullUpdateCollection(collection entity.ArtCollection) (entity.ArtCollection, error) {
	collection, err := s.artCollectionRepository.FullUpdateCollection(collection)
	if err != nil {
		return entity.ArtCollection{}, err
	}
	return collection, nil
}
