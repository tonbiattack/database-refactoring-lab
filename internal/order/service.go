package order

import "context"

type Service struct {
	repository Repository
	readMode   ReadMode
	writeMode  WriteMode
}

func NewService(repository Repository, readMode ReadMode, writeMode WriteMode) *Service {
	return &Service{
		repository: repository,
		readMode:   readMode,
		writeMode:  writeMode,
	}
}

func (s *Service) Get(ctx context.Context, id int64) (OrderView, error) {
	order, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return OrderView{}, err
	}

	view := OrderView{
		ID:           order.ID,
		CustomerNote: order.CustomerNote,
	}
	if order.Status != nil {
		view.Status = *order.Status
		return view, nil
	}
	if s.readMode == ReadModeFallback {
		if status, ok := StatusFromLegacy(order.StatusCode); ok {
			view.Status = status.Code
		}
	}
	return view, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id int64, statusCode string) error {
	status, err := ParseStatus(statusCode)
	if err != nil {
		return err
	}
	return s.repository.UpdateStatus(ctx, id, status, s.writeMode)
}
