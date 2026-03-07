package service

// service层 实现具体的业务规则,计算逻辑,状态流转判断
// 区别于 DAO层的是, DAO层是怎么从数据库取数据
import (
	"errors"
	"log"
	"time"
	"user-management/internal/model"
	"user-management/internal/repository"

	"regexp"
)

const SUPERUSER = "admin"
const COMMON = "user"

type UserService struct {
	// 就是sql.DB
	Repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{Repo: repo}
}

// 预编译全局变量，提高性能（避免每次调用都重新编译）
var alphaNumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// Login 处理登录校验
func (s *UserService) Login(username, password string) (*model.User, error) {
	// 先判断输入账号格式
	ok, err := s.IsValidUsername(username)
	if !ok {
		return nil, err
	}
	// 判断密码格式
	ok, err = s.IsValidPassword(password)
	if !ok {
		return nil, err
	}
	// 判断账号是否存在
	okU, _ := s.IsExistUsername(username)

	// 判断密码是否正确
	okP, _ := s.VerifyPassword(username, password)

	if !okU || !okP {
		return nil, errors.New("用户名或者密码输入错误")
	}

	// 最后返回这个用户结构体
	result, _ := s.Repo.GetUserByUsername(username)

	return result, nil
}

// Register 处理注册校验
func (s *UserService) Register(username, password, email string) (*model.User, error) {

	// 账号是否合法
	ok, err := s.IsValidUsername(username)
	if !ok {
		return nil, err
	}
	// 密码是否合法
	ok, err = s.IsValidPassword(password)
	if !ok {
		return nil, err
	}
	// 账号是否已存在
	ok, err = s.IsExistUsername(username)
	if ok {
		// 这里是注册---账号存在就退出
		log.Printf("用户名已存在: %s", username)
		return nil, err
	}
	// 创建用户结构体,并插入---密码暂时明文存储
	now := time.Now()
	u := model.User{
		Username:  username,
		Password:  password,
		Email:     email,
		AvatarURL: "",
		Role:      COMMON,
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, CreateErr := s.Repo.CreateUser(&u)
	if CreateErr != nil {
		return nil, errors.New("注册失败")
	}
	return &u, nil
}

// IsValidUsername 验证用户格式
func (s *UserService) IsValidUsername(username string) (bool, error) {
	// 判断长度
	if len(username) < 6 || len(username) > 13 {
		return false, errors.New("用户名长度不能小于6位,不能大于13位")
	}
	// 使用正则表达式：必须只包含字母和数字
	if !alphaNumRegex.MatchString(username) {
		return false, errors.New("用户名格式错误,只能由字母和数字组合")
	}
	return true, nil
}

// IsExistUsername 判断用户名是否存在
func (s *UserService) IsExistUsername(username string) (bool, error) {
	// 不看返回的user,就看有没有这个用户
	_, err := s.Repo.GetUserByUsername(username)

	// 如果 err 消息包含"用户不存在"，说明用户不存在
	if err != nil && err.Error() == "用户不存在" {
		return false, nil
	}

	// 如果 err != nil，说明查询失败
	if err != nil {
		return false, err
	}
	// 用户存在
	return true, nil
}

// IsValidPassword 验证密码格式
func (s *UserService) IsValidPassword(password string) (bool, error) {
	// 判断长度
	if len(password) < 6 || len(password) > 13 {
		return false, errors.New("密码长度不能小于6位,不能大于13位")
	}
	// 使用正则表达式：必须只包含字母和数字
	if !alphaNumRegex.MatchString(password) {
		return false, errors.New("密码格式错误,只能由字母和数字组合")
	}
	return true, nil
}

// VerifyPassword 验证密码是否正确
func (s *UserService) VerifyPassword(username, password string) (bool, error) {
	u, err := s.Repo.GetUserByUsername(username)

	if err != nil {
		// 日志在后端打印,给用户不返回err
		log.Printf("获取用户结构体失败:%v", err)
		return false, errors.New("用户名或密码错误")
	}

	if u.Password != password {
		return false, errors.New("用户名或密码错误")
	}
	return true, nil
}

// Identify 身份识别,获取用户的身份
func (s *UserService) Identify(username string) (string, error) {
	u, err := s.Repo.GetUserByUsername(username)
	if err != nil {
		return "", errors.New("获取用户身份失败")
	}

	if u.Role == SUPERUSER {
		return "admin", nil
	}
	if u.Role == COMMON {
		return "user", nil
	}
	return "", errors.New("未知角色")
}

func (s *UserService) GetUserList(page, pageSize int) ([]model.User, int, int, error) {
	offset := (page - 1) * pageSize
	// 获取所有用户信息
	users, SqlErr := s.Repo.GetAllList(offset, pageSize)
	if SqlErr != nil {
		return nil, 0, 0, SqlErr
	}

	// 设置默认头像
	for i := range users {
		if users[i].AvatarURL == "" {
			users[i].AvatarURL = "/static/images/default-avatar.png"
		}
	}

	// 获取用户数量
	totalCount, err := s.Repo.GetUserCount()
	if err != nil {
		return nil, 0, 0, err
	}
	totalPages := (totalCount + pageSize - 1) / pageSize
	return users, totalCount, totalPages, nil

}
