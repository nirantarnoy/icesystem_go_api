package service

import (
	"github.com/mashingan/smapping"
	"tarlek.com/icesystem/dto"
	"tarlek.com/icesystem/entity"
	"tarlek.com/icesystem/repository"
)

type PosCloseService interface {
	PosClose(poscloseDto dto.PosCloseDataDto) entity.PosBill
}

type poscloseService struct {
	posCloseRepository repository.PosCloseRepository
}

// PosClose implements PosCloseService
func (posSv *poscloseService) PosClose(poscloseDto dto.PosCloseDataDto) entity.PosBill {
	posData := entity.PosCloseData{}
	err := smapping.FillStruct(&posData, smapping.MapFields(&poscloseDto))
	if err != nil {
		print(err)
	}

	return posSv.posCloseRepository.CloseOrder(posData)
}

func NewPosCloseService(repo repository.PosCloseRepository) PosCloseService {
	return &poscloseService{posCloseRepository: repo}
}
