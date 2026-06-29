package service

import (
	"errors"

	"go-admin/internal/common"
	"go-admin/internal/module/system/model"
	"go-admin/internal/module/system/repository"

	"gorm.io/gorm"
)

type PostService interface {
	Create(name, code string, sort int, status int8, operatorID uint) error
	Update(id uint, name, code string, sort int, status int8, operatorID uint) error
	Delete(id uint) error
	FindByID(id uint) (interface{}, error)
	FindAll() ([]model.SysPost, error)
	FindList(name string, status *int8, page, pageSize int) ([]interface{}, int64, error)
	UpdateStatus(id uint, status int8) error
}

type postService struct {
	postRepo repository.PostRepository
}

func NewPostService() PostService {
	return &postService{
		postRepo: repository.NewPostRepository(),
	}
}

func (s *postService) Create(name, code string, sort int, status int8, operatorID uint) error {
	if s.postRepo.CountByCode(code, 0) > 0 {
		return errors.New("岗位编码已存在")
	}

	post := &model.SysPost{
		BaseModel: common.BaseModel{
			CreateBy: operatorID,
			UpdateBy: operatorID,
		},
		Code:   code,
		Name:   name,
		Sort:   sort,
		Status: status,
	}

	return s.postRepo.Create(post)
}

func (s *postService) Update(id uint, name, code string, sort int, status int8, operatorID uint) error {
	if s.postRepo.CountByCode(code, id) > 0 {
		return errors.New("岗位编码已存在")
	}

	post, err := s.postRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("岗位不存在")
		}
		return err
	}

	post.Name = name
	post.Code = code
	post.Sort = sort
	post.Status = status
	post.UpdateBy = operatorID

	return s.postRepo.Update(post)
}

func (s *postService) Delete(id uint) error {
	return s.postRepo.Delete(id)
}

func (s *postService) FindByID(id uint) (interface{}, error) {
	return s.postRepo.FindByID(id)
}

func (s *postService) FindAll() ([]model.SysPost, error) {
	return s.postRepo.FindAll()
}

func (s *postService) FindList(name string, status *int8, page, pageSize int) ([]interface{}, int64, error) {
	posts, total, err := s.postRepo.FindList(name, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]interface{}, len(posts))
	for i, p := range posts {
		result[i] = p
	}
	return result, total, nil
}

func (s *postService) UpdateStatus(id uint, status int8) error {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return err
	}
	post.Status = status
	return s.postRepo.Update(post)
}
