package utils

// Authentication errors
const (
	ErrAuthenticationKeyNotFound = "authentication_key_not_found"
	ErrUnauthorized              = "unauthorized"
	ErrTokenExpired              = "token_expired"
)

// Request errors
const (
	ErrBadRequest = "bad_request"
	ErrUserIDNotFound = "user_id_not_found"
)

// User-related errors
const (
	ErrInvalidUsernameOrEmail = "invalid_username_or_email"
	ErrInvalidPassword        = "invalid_password"
	ErrEmailAlreadyUsed       = "email_already_used"
	ErrUsernameAlreadyUsed    = "username_already_used"
	ErrEmailNotVerify         = "email_not_verify"
)

// Courses-releated errors
const (
	ErrDuplicateCourseStage = "duplicate_courses_stage"
	ErrImageRequired = "image_required"
	ErrMissingCourseID = "missing_course_id"
	ErrCourseNotExist = "course_not_exist"
	ErrImageTooLarge = "image_too_large"
	ErrOpeningImage = "opening_image_failed"
	ErrReadingImage = "reading_image_failed"
	ErrInvalidImage  = "invalid_image"
	ErrResetFilePointer  = "reset_file_pointer_failed"
	ErrS3UploadFailed = "s3_upload_failed"
)

// Database errors
const (
	ErrSaveData = "error_save_data"
	ErrGetData  = "error_get_data"
)

// Internal errors
const (
	ErrHashData        = "hash_data_failed"
	ErrParseFile       = "template_parse_failed"
	ErrSendEmail       = "send_email_failed"
	ErrGenerateSession = "generate_session_failed"
	ErrParseData       = "parse_data_failed"
	ErrGenerateToken   = "generate_token_failed"
	ErrExecuteTemplate = "execute_template_failed"
	ErrStoreRedis      = "store_redis_failed"
)

// Gitea errors
const (
	ErrCreateNewCourse = "create_new_course_failed"
	ErrSaveCourseFile = "save_new_course_file_failed"
)