package service_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/service"

	"github.com/stretchr/testify/suite"
)

type mockAuthorRepository struct {
	mock.Mock
}

func (m *mockAuthorRepository) GetAllAuthors() ([]entity.Author, error) {
	args := m.Called()
	return args.Get(0).([]entity.Author), args.Error(1)
}
func (m *mockAuthorRepository) CreateAuthor(author entity.Author) (entity.Author, error) {
	args := m.Called()
	return args.Get(0).(entity.Author), args.Error(1)
}
func (m *mockAuthorRepository) GetAuthorByID(id uint) (entity.Author, error) {
	args := m.Called()
	return args.Get(0).(entity.Author), args.Error(1)
}
func (m *mockAuthorRepository) UpdateAuthor(author entity.Author) (entity.Author, error) {
	args := m.Called()
	return args.Get(0).(entity.Author), args.Error(1)
}
func (m *mockAuthorRepository) DeleteAuthor(id uint) error {
	args := m.Called()
	return args.Error(0)

}
func (m *mockAuthorRepository) PartialUpdateAuthor(id uint, kwargs map[string]interface{}) (entity.Author, error) {
	args := m.Called()
	return args.Get(0).(entity.Author), args.Error(1)
}
func (m *mockAuthorRepository) FullUpdateAuthor(author entity.Author) (entity.Author, error) {
	args := m.Called()
	return args.Get(0).(entity.Author), args.Error(1)
}

type AuthorServiceTestSuite struct {
	suite.Suite
	service      service.AuthorService
	two_authors  []entity.Author
	five_authors []entity.Author
	one_author   entity.Author
	mockRepo     *mockAuthorRepository
}

func (suite *AuthorServiceTestSuite) SetupTest() {

	suite.two_authors = []entity.Author{
		{Name: entity.TranslatedText{EN: "Author 1", RU: "Автор 1", ES: "Autor 1"}},
		{Name: entity.TranslatedText{EN: "Author 2", RU: "Автор 2", ES: "Autor 2"}},
	}
	suite.five_authors = []entity.Author{
		{Name: entity.TranslatedText{EN: "Author 1", RU: "Автор 1", ES: "Autor 1"}},
		{Name: entity.TranslatedText{EN: "Author 2", RU: "Автор 2", ES: "Autor 2"}},
		{Name: entity.TranslatedText{EN: "Author 3", RU: "Автор 3", ES: "Autor 3"}},
		{Name: entity.TranslatedText{EN: "Author 4", RU: "Автор 4", ES: "Autor 4"}},
		{Name: entity.TranslatedText{EN: "Author 5", RU: "Автор 5", ES: "Autor 5"}},
	}
	suite.one_author = entity.Author{
		Name: entity.TranslatedText{EN: "Author 1", RU: "Автор 1", ES: "Autor 1"},
	}
	suite.mockRepo = new(mockAuthorRepository)
	suite.service = service.NewAuthorService(suite.mockRepo)
}

func (suite *AuthorServiceTestSuite) TestGetAllAuthorsSuccess() {
	suite.mockRepo.On("GetAllAuthors").Return(suite.two_authors, nil)
	authors, err := suite.service.GetAllAuthors()
	suite.NoError(err)

	suite.Equal(suite.two_authors, authors)
	suite.mockRepo.AssertExpectations(suite.T())
	// suite.Equal(1, err2)

}
func (suite *AuthorServiceTestSuite) TestGetAllAuthorsError() {
	suite.mockRepo.On("GetAllAuthors").Return([]entity.Author{}, errors.New("error"))
	authors, err := suite.service.GetAllAuthors()
	suite.Error(err)

	suite.Equal(authors, []entity.Author{})
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestCreateAuthorSuccess() {
	suite.mockRepo.On("CreateAuthor").Return(suite.one_author, nil)

	author, err := suite.service.CreateAuthor(entity.Author{
		Name: entity.TranslatedText{EN: "Author 1", RU: "Автор 1", ES: "Autor 1"},
	})
	suite.NoError(err)

	suite.Equal(suite.one_author, author)
	suite.mockRepo.AssertExpectations(suite.T())
}
func (suite *AuthorServiceTestSuite) TestCreateAuthorError() {
	suite.mockRepo.On("CreateAuthor").Return(suite.one_author, errors.New("error"))

	author, err := suite.service.CreateAuthor(suite.one_author)
	suite.Error(err)

	suite.Equal(suite.one_author, author)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestGetAuthorByIDSuccess() {
	suite.mockRepo.On("GetAuthorByID").Return(suite.one_author, nil)

	author, err := suite.service.GetAuthorByID(1)
	suite.NoError(err)

	suite.Equal(suite.one_author, author)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestGetAuthorByIDError() {
	suite.mockRepo.On("GetAuthorByID").Return(entity.Author{}, errors.New("error"))

	author, err := suite.service.GetAuthorByID(1)
	suite.Error(err)

	suite.Equal(entity.Author{}, author)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestUpdateAuthorSuccess() {
	suite.mockRepo.On("UpdateAuthor").Return(suite.one_author, nil)

	author, err := suite.service.UpdateAuthor(suite.one_author)
	suite.NoError(err)

	suite.Equal(suite.one_author, author)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestUpdateAuthorError() {
	suite.mockRepo.On("UpdateAuthor").Return(entity.Author{}, errors.New("error"))

	author, err := suite.service.UpdateAuthor(suite.one_author)
	suite.Error(err)

	suite.Equal(entity.Author{}, author)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestDeleteAuthorSuccess() {
	suite.mockRepo.On("DeleteAuthor").Return(nil)

	err := suite.service.DeleteAuthor(1)
	suite.NoError(err)

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestDeleteAuthorError() {
	suite.mockRepo.On("DeleteAuthor").Return(errors.New("error"))

	err := suite.service.DeleteAuthor(1)
	suite.Error(err)

	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestPartialUpdateAuthorSuccess() {
	suite.mockRepo.On("PartialUpdateAuthor").Return(suite.one_author, nil)

	author, err := suite.service.PartialUpdateAuthor(1, map[string]interface{}{"Name": "Author 1"})
	suite.NoError(err)

	suite.Equal(suite.one_author, author)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestPartialUpdateAuthorError() {
	suite.mockRepo.On("PartialUpdateAuthor").Return(entity.Author{}, errors.New("error"))

	author, err := suite.service.PartialUpdateAuthor(1, map[string]interface{}{"Name": "Author 1"})
	suite.Error(err)

	suite.Equal(entity.Author{}, author)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestFullUpdateAuthorSuccess() {
	suite.mockRepo.On("FullUpdateAuthor").Return(suite.one_author, nil)

	author, err := suite.service.FullUpdateAuthor(suite.one_author)
	suite.NoError(err)

	suite.Equal(suite.one_author, author)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuthorServiceTestSuite) TestFullUpdateAuthorError() {
	suite.mockRepo.On("FullUpdateAuthor").Return(entity.Author{}, errors.New("error"))

	author, err := suite.service.FullUpdateAuthor(suite.one_author)
	suite.Error(err)

	suite.Equal(entity.Author{}, author)
	suite.mockRepo.AssertExpectations(suite.T())
}

func TestAuthorServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorServiceTestSuite))
}
