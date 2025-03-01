package service

type InterviewsService struct {
	repo *repository.Repository
	ai AIInterviews
}

type AIInterviews interface {
	MakeInterview(stack string, difficultyLevel string) (AIInterview, error)
}

func (s *InterviewsService) MakeInterview(stack string, difficultyLevel string, userId int) (int, error) {
	
}