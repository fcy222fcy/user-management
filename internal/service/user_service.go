package service

// service层 实现具体的业务规则,计算逻辑,状态流转判断
// 区别于 DAO层的是, DAO层是怎么从数据库取数据
import (
	"errors"
	"fmt"
	"log"
	"strings"
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
	if err := s.IsValidUsername(username); err != nil {
		return nil, err
	}
	// 判断密码格式
	if err := s.IsValidPassword(password); err != nil {
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
	if err := s.IsValidUsername(username); err != nil {
		return nil, err
	}
	// 密码是否合法
	if err := s.IsValidPassword(password); err != nil {
		return nil, err
	}
	// 账号是否已存在
	exists, err := s.IsExistUsername(username)
	if err != nil {
		// 这里是注册---账号存在就退出
		return nil, fmt.Errorf("查询用户名失败:%w", err)
	}
	if exists {
		log.Printf("用户名已存在: %s", username)
		return nil, errors.New("用户名已存在")
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
		return nil, fmt.Errorf("注册失败:%w", err)
	}
	return &u, nil
}

// IsValidUsername 验证用户格式
func (s *UserService) IsValidUsername(username string) error {
	// 判断长度
	if len(username) < 6 || len(username) > 13 {
		return errors.New("用户名长度不能小于6位,不能大于13位")
	}
	// 使用正则表达式：必须只包含字母和数字
	if !alphaNumRegex.MatchString(username) {
		return errors.New("用户名格式错误,只能由字母和数字组合")
	}
	return nil
}

// IsExistUsername 判断用户名是否存在
func (s *UserService) IsExistUsername(username string) (bool, error) {
	_, err := s.Repo.GetUserByUsername(username)

	if err != nil {
		if strings.Contains(err.Error(), "用户不存在") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsValidPassword 验证密码格式
func (s *UserService) IsValidPassword(password string) error {
	// 判断长度
	if len(password) < 6 || len(password) > 13 {
		return errors.New("密码长度不能小于6位,不能大于13位")
	}
	// 使用正则表达式：必须只包含字母和数字
	if !alphaNumRegex.MatchString(password) {
		return errors.New("密码格式错误,只能由字母和数字组合")
	}
	return nil
}

// VerifyPassword 验证密码是否正确
func (s *UserService) VerifyPassword(username, password string) (bool, error) {
	u, err := s.Repo.GetUserByUsername(username)

	if err != nil {
		// 日志在后端打印,给用户不返回err
		if errors.Is(err, errors.New("用户不存在")) {
			return false, errors.New("用户名或密码错误")
		}
		log.Printf("获取用户结构体失败:%v", err)
		return false, errors.New("系统错误")
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

// GetUserList 获取用户列表,带分页
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

// SearchUsers 搜索用户，支持用户名模糊搜索和状态筛选
func (s *UserService) SearchUsers(keyword string, status int, page, pageSize int) ([]model.User, int, int, error) {
	offset := (page - 1) * pageSize

	// 获取搜索的用户列表
	users, err := s.Repo.SearchUsers(keyword, status, offset, pageSize)
	if err != nil {
		return nil, 0, 0, err
	}

	// 设置默认头像
	for i := range users {
		if users[i].AvatarURL == "" {
			users[i].AvatarURL = "/static/images/default-avatar.png"
		}
	}

	// 获取搜索的总数
	totalCount, err := s.Repo.GetSearchUsersCount(keyword, status)
	if err != nil {
		return nil, 0, 0, err
	}
	totalPages := (totalCount + pageSize - 1) / pageSize
	return users, totalCount, totalPages, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(currentUserRole string, username, password, email, avatarURL, role string, status int) (int64, error) {
	// 权限检查
	if currentUserRole != "admin" {
		return 0, errors.New("没有权限添加用户")
	}
	// 用户名格式校验
	if err := s.IsValidUsername(username); err != nil {
		return 0, err
	}
	// 密码格式校验
	if err := s.IsValidPassword(password); err != nil {
		return 0, err
	}
	// 用户名重复检查
	exists, err := s.IsExistUsername(username)
	if err != nil {
		return 0, fmt.Errorf("查询用户名失败 %w", err)
	}
	if exists {
		return 0, errors.New("用户名已存在")
	}
	// 创建用户结构体
	u := model.User{
		Username:  username,
		Password:  password,
		Email:     email,
		AvatarURL: avatarURL,
		Role:      role,
		Status:    status,
	}
	// 插入数据库
	id, CreateErr := s.Repo.CreateUser(&u)
	if CreateErr != nil {
		return 0, CreateErr
	}
	return id, nil
}

// DeleteUser 根据ID删除用户
func (s *UserService) DeleteUser(currentUserID int64, currentUserRole string, targetUserID int64) error {

	// 权限检查
	if currentUserRole != "admin" {
		return errors.New("没有权限删除用户")
	}

	// 不能删除自己
	if currentUserID == targetUserID {
		return errors.New("不能删除自己")
	}

	// 检查要删除的用户是否存在
	targetUser, err := s.Repo.GetUserByID(targetUserID)
	if err != nil || targetUser == nil {
		return errors.New("用户不存在")
	}
	if targetUser.Role == "admin" {
		return errors.New("不能删除管理员")
	}

	// 执行删除
	return s.Repo.DeleteUser(targetUserID)
}

// GetUserByID 根据ID获取用户信息
func (s *UserService) GetUserByID(id int64) (*model.User, error) {
	return s.Repo.GetUserByID(id)
}

func (s *UserService) GetDashboardStats() (*model.DashboardStats, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
	startOfPrevMonth := startOfMonth.AddDate(0, -1, 0)

	totalUsers, err := s.Repo.GetUserCount()
	if err != nil {
		return nil, err
	}

	totalUsersPrev, err := s.Repo.GetUserCountBefore(startOfMonth)
	if err != nil {
		return nil, err
	}

	totalUsersPrevMonth, err := s.Repo.GetUserCountBefore(startOfPrevMonth)
	if err != nil {
		return nil, err
	}

	monthLogins, err := s.Repo.GetLoginCountInRange(startOfMonth, startOfNextMonth)
	if err != nil {
		return nil, err
	}

	prevMonthLogins, err := s.Repo.GetLoginCountInRange(startOfPrevMonth, startOfMonth)
	if err != nil {
		return nil, err
	}

	// 从login_logs统计删除用户（注销用户）
	deletedUsers, err := s.Repo.GetDeletedUsersCount()
	if err != nil {
		return nil, err
	}

	deletedThisMonth, err := s.Repo.GetDeletedUsersCountInRange(startOfMonth, startOfNextMonth)
	if err != nil {
		return nil, err
	}

	deletedPrevMonth, err := s.Repo.GetDeletedUsersCountInRange(startOfPrevMonth, startOfMonth)
	if err != nil {
		return nil, err
	}

	trendMap, err := s.Repo.GetDailyLogins(startOfPrevMonth, startOfNextMonth)
	if err != nil {
		return nil, err
	}

	trend := make([]model.TrendPoint, 0, 62)
	for d := startOfPrevMonth; d.Before(startOfNextMonth); d = d.AddDate(0, 0, 1) {
		day := d.Format("2006-01-02")
		trend = append(trend, model.TrendPoint{
			Date:  day,
			Count: trendMap[day],
		})
	}

	stats := &model.DashboardStats{
		TotalUsers:       totalUsers,
		UserGrowthRate:   calcGrowthRate(totalUsersPrev, totalUsersPrevMonth),
		MonthLogins:      monthLogins,
		LoginGrowthRate:  calcGrowthRate(monthLogins, prevMonthLogins),
		DeactivatedUsers: deletedUsers,
		DeactivatedRate:  calcGrowthRate(deletedThisMonth, deletedPrevMonth),
		LoginTrend:       trend,
	}

	return stats, nil
}

func calcGrowthRate(current, previous int) int {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return int((float64(current-previous) / float64(previous)) * 100)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(currentUserID int64, currentUserRole string, targetUserID int64, username, password, email string, status int, avatarURL string, role string) error {

	// 权限检查：只有管理员可以修改其他用户，普通用户只能修改自己的信息
	if currentUserRole != "admin" && currentUserID != targetUserID {
		return errors.New("没有权限修改其他用户信息")
	}

	// 普通用户不能修改角色和状态
	if currentUserRole != "admin" {
		role = ""   // 保持原角色
		status = -1 // 标记为不更新状态
	}

	// 1 查询所修改的用户是否存在
	u, err := s.GetUserByID(targetUserID)
	if err != nil {
		return fmt.Errorf("查找所修改的用户失败: %w", err)
	}

	if u == nil {
		return errors.New("所修改的用户不存在")
	}

	// 2 用户名格式校验（如果填写了）
	if username != "" {
		// 判断是否合法
		if err := s.IsValidUsername(username); err != nil {
			return err
		}

		// 用户名重复检查
		exists, err := s.IsExistUsername(username)
		if err != nil {
			return fmt.Errorf("查询用户名失败: %w", err)
		}

		if exists && username != u.Username {
			return errors.New("用户名已存在")
		}
	} else {
		username = u.Username
	}

	// 3 密码格式校验(如果填写了)
	if password != "" {
		if err := s.IsValidPassword(password); err != nil {
			return err
		}
	} else {
		// 密码为空,就填写原密码
		password = u.Password
	}

	// 4 email 空值填充
	if email == "" {
		email = u.Email
	}

	// 5 角色空值填充
	if role == "" {
		role = u.Role
	}

	// 6 状态填充
	if status == -1 {
		status = u.Status
	}

	// avatarURL如果是空的话,就代表没有更新,就不用管

	// 7 更新数据库
	return s.Repo.UpdateUser(targetUserID, username, password, email, status, avatarURL, role)
}
