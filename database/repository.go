// Package database - repository untuk operasi database auto promote
package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Repository interface untuk operasi database
type Repository interface {
	// Auto Promote Groups
	GetAutoPromoteGroup(groupJID string) (*AutoPromoteGroup, error)
	CreateAutoPromoteGroup(groupJID string) (*AutoPromoteGroup, error)
	UpdateAutoPromoteGroup(group *AutoPromoteGroup) error
	GetActiveGroups() ([]AutoPromoteGroup, error)
	
	// Promote Templates
	GetAllTemplates() ([]PromoteTemplate, error)
	GetActiveTemplates() ([]PromoteTemplate, error)
	GetTemplateByID(id int) (*PromoteTemplate, error)
	CreateTemplate(template *PromoteTemplate) error
	UpdateTemplate(template *PromoteTemplate) error
	DeleteTemplate(id int) error
	
	// Promote Logs
	CreateLog(log *PromoteLog) error
	GetLogsByGroup(groupJID string, limit int) ([]PromoteLog, error)
	
	// Stats
	UpdateStats(date string, totalGroups, totalMessages, successMessages, failedMessages int) error
	GetStats(date string) (*PromoteStats, error)
	
	// Learning Bot methods
	// Learning Groups
	CreateLearningGroup(group *LearningGroup) error
	GetLearningGroup(groupJID string) (*LearningGroup, error)
	GetAllLearningGroups() ([]LearningGroup, error)
	UpdateLearningGroup(group *LearningGroup) error
	DeleteLearningGroup(groupJID string) error
	ToggleLearningGroup(groupJID string, isActive bool) error
	
	// Learning Commands
	CreateLearningCommand(cmd *LearningCommand) error
	GetLearningCommand(command string) (*LearningCommand, error)
	GetAllLearningCommands() ([]LearningCommand, error)
	GetLearningCommandsByCategory(category string) ([]LearningCommand, error)
	UpdateLearningCommand(cmd *LearningCommand) error
	DeleteLearningCommand(command string) error
	IncrementCommandUsage(command string) error
	
	// Auto Responses
	CreateAutoResponse(response *AutoResponse) error
	GetAutoResponse(keyword string) (*AutoResponse, error)
	GetAutoResponsesByKeyword(text string) ([]AutoResponse, error)
	GetAllAutoResponses() ([]AutoResponse, error)
	UpdateAutoResponse(response *AutoResponse) error
	DeleteAutoResponse(keyword string) error
	IncrementAutoResponseUsage(keyword string) error
	
	// Command Usage Logs
	LogCommandUsage(log *CommandUsageLog) error
	GetCommandUsageLogs(limit int) ([]CommandUsageLog, error)
	GetCommandUsageStats(days int) (map[string]int, error)

	// Forbidden Words
	CreateForbiddenWord(word *ForbiddenWord) error
	GetForbiddenWordsByGroup(groupJID string) ([]ForbiddenWord, error)
	DeleteForbiddenWord(id int) error
}

// SQLiteRepository implementasi repository untuk SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository membuat repository baru
func NewSQLiteRepository(db *sql.DB) Repository {
	return &SQLiteRepository{db: db}
}

// === AUTO PROMOTE GROUPS ===

func (r *SQLiteRepository) GetAutoPromoteGroup(groupJID string) (*AutoPromoteGroup, error) {
	query := `SELECT id, group_jid, is_active, started_at, last_promote_at, created_at, updated_at 
			  FROM auto_promote_groups WHERE group_jid = ?`
	
	row := r.db.QueryRow(query, groupJID)
	
	var group AutoPromoteGroup
	var startedAt, lastPromoteAt sql.NullTime
	
	err := row.Scan(&group.ID, &group.GroupJID, &group.IsActive, 
		&startedAt, &lastPromoteAt, &group.CreatedAt, &group.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Group tidak ditemukan
		}
		return nil, err
	}
	
	if startedAt.Valid {
		group.StartedAt = &startedAt.Time
	}
	if lastPromoteAt.Valid {
		group.LastPromoteAt = &lastPromoteAt.Time
	}
	
	return &group, nil
}

func (r *SQLiteRepository) CreateAutoPromoteGroup(groupJID string) (*AutoPromoteGroup, error) {
	query := `INSERT INTO auto_promote_groups (group_jid, is_active, created_at, updated_at) 
			  VALUES (?, ?, ?, ?)`
	
	now := time.Now()
	result, err := r.db.Exec(query, groupJID, false, now, now)
	if err != nil {
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	return &AutoPromoteGroup{
		ID:        int(id),
		GroupJID:  groupJID,
		IsActive:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *SQLiteRepository) UpdateAutoPromoteGroup(group *AutoPromoteGroup) error {
	query := `UPDATE auto_promote_groups 
			  SET is_active = ?, started_at = ?, last_promote_at = ?, updated_at = ? 
			  WHERE id = ?`
	
	group.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(query, group.IsActive, group.StartedAt, 
		group.LastPromoteAt, group.UpdatedAt, group.ID)
	
	return err
}

func (r *SQLiteRepository) GetActiveGroups() ([]AutoPromoteGroup, error) {
	query := `SELECT id, group_jid, is_active, started_at, last_promote_at, created_at, updated_at 
			  FROM auto_promote_groups WHERE is_active = true`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var groups []AutoPromoteGroup
	
	for rows.Next() {
		var group AutoPromoteGroup
		var startedAt, lastPromoteAt sql.NullTime
		
		err := rows.Scan(&group.ID, &group.GroupJID, &group.IsActive,
			&startedAt, &lastPromoteAt, &group.CreatedAt, &group.UpdatedAt)
		if err != nil {
			return nil, err
		}
		
		if startedAt.Valid {
			group.StartedAt = &startedAt.Time
		}
		if lastPromoteAt.Valid {
			group.LastPromoteAt = &lastPromoteAt.Time
		}
		
		groups = append(groups, group)
	}
	
	return groups, nil
}

// === PROMOTE TEMPLATES ===

func (r *SQLiteRepository) GetAllTemplates() ([]PromoteTemplate, error) {
	query := `SELECT id, title, content, category, is_active, created_at, updated_at 
			  FROM promote_templates ORDER BY created_at DESC`
	
	return r.queryTemplates(query)
}

func (r *SQLiteRepository) GetActiveTemplates() ([]PromoteTemplate, error) {
	query := `SELECT id, title, content, category, is_active, created_at, updated_at 
			  FROM promote_templates WHERE is_active = true ORDER BY created_at DESC`
	
	return r.queryTemplates(query)
}

func (r *SQLiteRepository) queryTemplates(query string, args ...interface{}) ([]PromoteTemplate, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var templates []PromoteTemplate
	
	for rows.Next() {
		var template PromoteTemplate
		err := rows.Scan(&template.ID, &template.Title, &template.Content,
			&template.Category, &template.IsActive, &template.CreatedAt, &template.UpdatedAt)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}
	
	return templates, nil
}

func (r *SQLiteRepository) GetTemplateByID(id int) (*PromoteTemplate, error) {
	query := `SELECT id, title, content, category, is_active, created_at, updated_at 
			  FROM promote_templates WHERE id = ?`
	
	row := r.db.QueryRow(query, id)
	
	var template PromoteTemplate
	err := row.Scan(&template.ID, &template.Title, &template.Content,
		&template.Category, &template.IsActive, &template.CreatedAt, &template.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &template, nil
}

func (r *SQLiteRepository) CreateTemplate(template *PromoteTemplate) error {
	query := `INSERT INTO promote_templates (title, content, category, is_active, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	
	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now
	
	result, err := r.db.Exec(query, template.Title, template.Content, 
		template.Category, template.IsActive, template.CreatedAt, template.UpdatedAt)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	template.ID = int(id)
	return nil
}

func (r *SQLiteRepository) UpdateTemplate(template *PromoteTemplate) error {
	query := `UPDATE promote_templates 
			  SET title = ?, content = ?, category = ?, is_active = ?, updated_at = ? 
			  WHERE id = ?`
	
	template.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(query, template.Title, template.Content, 
		template.Category, template.IsActive, template.UpdatedAt, template.ID)
	
	return err
}

func (r *SQLiteRepository) DeleteTemplate(id int) error {
	query := `DELETE FROM promote_templates WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// === PROMOTE LOGS ===

func (r *SQLiteRepository) CreateLog(log *PromoteLog) error {
	query := `INSERT INTO promote_logs (group_jid, template_id, content, sent_at, success, error_msg) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.Exec(query, log.GroupJID, log.TemplateID, 
		log.Content, log.SentAt, log.Success, log.ErrorMsg)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	log.ID = int(id)
	return nil
}

func (r *SQLiteRepository) GetLogsByGroup(groupJID string, limit int) ([]PromoteLog, error) {
	query := `SELECT id, group_jid, template_id, content, sent_at, success, error_msg 
			  FROM promote_logs WHERE group_jid = ? 
			  ORDER BY sent_at DESC LIMIT ?`
	
	rows, err := r.db.Query(query, groupJID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var logs []PromoteLog
	
	for rows.Next() {
		var log PromoteLog
		var errorMsg sql.NullString
		
		err := rows.Scan(&log.ID, &log.GroupJID, &log.TemplateID,
			&log.Content, &log.SentAt, &log.Success, &errorMsg)
		if err != nil {
			return nil, err
		}
		
		if errorMsg.Valid {
			log.ErrorMsg = &errorMsg.String
		}
		
		logs = append(logs, log)
	}
	
	return logs, nil
}

// === STATS ===

func (r *SQLiteRepository) UpdateStats(date string, totalGroups, totalMessages, successMessages, failedMessages int) error {
	query := `INSERT OR REPLACE INTO promote_stats 
			  (date, total_groups, total_messages, success_messages, failed_messages, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err := r.db.Exec(query, date, totalGroups, totalMessages, 
		successMessages, failedMessages, time.Now())
	
	return err
}

func (r *SQLiteRepository) GetStats(date string) (*PromoteStats, error) {
	query := `SELECT id, date, total_groups, total_messages, success_messages, failed_messages, created_at 
			  FROM promote_stats WHERE date = ?`
	
	row := r.db.QueryRow(query, date)
	
	var stats PromoteStats
	err := row.Scan(&stats.ID, &stats.Date, &stats.TotalGroups,
		&stats.TotalMessages, &stats.SuccessMessages, &stats.FailedMessages, &stats.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &stats, nil
}

// ===============================
// LEARNING BOT REPOSITORY METHODS
// ===============================

// === LEARNING GROUPS ===

func (r *SQLiteRepository) CreateLearningGroup(group *LearningGroup) error {
	query := `INSERT INTO learning_groups (group_jid, group_name, is_active, description, created_by, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	now := time.Now()
	_, err := r.db.Exec(query, group.GroupJID, group.GroupName, group.IsActive, group.Description, 
		group.CreatedBy, now, now)
	
	return err
}

func (r *SQLiteRepository) GetLearningGroup(groupJID string) (*LearningGroup, error) {
	query := `SELECT id, group_jid, group_name, is_active, description, created_by, created_at, updated_at 
			  FROM learning_groups WHERE group_jid = ?`
	
	var group LearningGroup
	err := r.db.QueryRow(query, groupJID).Scan(
		&group.ID, &group.GroupJID, &group.GroupName, &group.IsActive, 
		&group.Description, &group.CreatedBy, &group.CreatedAt, &group.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &group, nil
}

func (r *SQLiteRepository) GetAllLearningGroups() ([]LearningGroup, error) {
	query := `SELECT id, group_jid, group_name, is_active, description, created_by, created_at, updated_at 
			  FROM learning_groups ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var groups []LearningGroup
	for rows.Next() {
		var group LearningGroup
		err := rows.Scan(&group.ID, &group.GroupJID, &group.GroupName, &group.IsActive,
			&group.Description, &group.CreatedBy, &group.CreatedAt, &group.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	
	return groups, nil
}

func (r *SQLiteRepository) UpdateLearningGroup(group *LearningGroup) error {
	query := `UPDATE learning_groups 
			  SET group_name = ?, is_active = ?, description = ?, updated_at = ? 
			  WHERE group_jid = ?`
	
	_, err := r.db.Exec(query, group.GroupName, group.IsActive, group.Description, 
		time.Now(), group.GroupJID)
	
	return err
}

func (r *SQLiteRepository) DeleteLearningGroup(groupJID string) error {
	query := `DELETE FROM learning_groups WHERE group_jid = ?`
	_, err := r.db.Exec(query, groupJID)
	return err
}

func (r *SQLiteRepository) ToggleLearningGroup(groupJID string, isActive bool) error {
	query := `UPDATE learning_groups SET is_active = ?, updated_at = ? WHERE group_jid = ?`
	_, err := r.db.Exec(query, isActive, time.Now(), groupJID)
	return err
}

// === LEARNING COMMANDS ===

func (r *SQLiteRepository) CreateLearningCommand(cmd *LearningCommand) error {
	query := `INSERT INTO learning_commands 
			  (command, title, description, response_type, text_content, media_file_path, caption, 
			   category, is_active, created_by, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	now := time.Now()
	_, err := r.db.Exec(query, cmd.Command, cmd.Title, cmd.Description, cmd.ResponseType,
		cmd.TextContent, cmd.MediaFilePath, cmd.Caption, cmd.Category, cmd.IsActive,
		cmd.CreatedBy, now, now)
	
	return err
}

func (r *SQLiteRepository) GetLearningCommand(command string) (*LearningCommand, error) {
	query := `SELECT id, command, title, description, response_type, text_content, 
			  media_file_path, caption, category, is_active, usage_count, created_by, created_at, updated_at 
			  FROM learning_commands WHERE command = ? AND is_active = 1`
	
	var cmd LearningCommand
	err := r.db.QueryRow(query, command).Scan(
		&cmd.ID, &cmd.Command, &cmd.Title, &cmd.Description, &cmd.ResponseType,
		&cmd.TextContent, &cmd.MediaFilePath, &cmd.Caption, &cmd.Category,
		&cmd.IsActive, &cmd.UsageCount, &cmd.CreatedBy, &cmd.CreatedAt, &cmd.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &cmd, nil
}

func (r *SQLiteRepository) GetAllLearningCommands() ([]LearningCommand, error) {
	query := `SELECT id, command, title, description, response_type, text_content, 
			  media_file_path, caption, category, is_active, usage_count, created_by, created_at, updated_at 
			  FROM learning_commands ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var commands []LearningCommand
	for rows.Next() {
		var cmd LearningCommand
		err := rows.Scan(&cmd.ID, &cmd.Command, &cmd.Title, &cmd.Description, &cmd.ResponseType,
			&cmd.TextContent, &cmd.MediaFilePath, &cmd.Caption, &cmd.Category,
			&cmd.IsActive, &cmd.UsageCount, &cmd.CreatedBy, &cmd.CreatedAt, &cmd.UpdatedAt)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	
	return commands, nil
}

func (r *SQLiteRepository) GetLearningCommandsByCategory(category string) ([]LearningCommand, error) {
	query := `SELECT id, command, title, description, response_type, text_content, 
			  media_file_path, caption, category, is_active, usage_count, created_by, created_at, updated_at 
			  FROM learning_commands WHERE category = ? AND is_active = 1 ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var commands []LearningCommand
	for rows.Next() {
		var cmd LearningCommand
		err := rows.Scan(&cmd.ID, &cmd.Command, &cmd.Title, &cmd.Description, &cmd.ResponseType,
			&cmd.TextContent, &cmd.MediaFilePath, &cmd.Caption, &cmd.Category,
			&cmd.IsActive, &cmd.UsageCount, &cmd.CreatedBy, &cmd.CreatedAt, &cmd.UpdatedAt)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	
	return commands, nil
}

func (r *SQLiteRepository) UpdateLearningCommand(cmd *LearningCommand) error {
	query := `UPDATE learning_commands 
			  SET title = ?, description = ?, response_type = ?, text_content = ?, 
			      media_file_path = ?, caption = ?, category = ?, is_active = ?, updated_at = ? 
			  WHERE command = ?`
	
	_, err := r.db.Exec(query, cmd.Title, cmd.Description, cmd.ResponseType,
		cmd.TextContent, cmd.MediaFilePath, cmd.Caption, cmd.Category, cmd.IsActive,
		time.Now(), cmd.Command)
	
	return err
}

func (r *SQLiteRepository) DeleteLearningCommand(command string) error {
	query := `DELETE FROM learning_commands WHERE command = ?`
	_, err := r.db.Exec(query, command)
	return err
}

func (r *SQLiteRepository) IncrementCommandUsage(command string) error {
	query := `UPDATE learning_commands SET usage_count = usage_count + 1 WHERE command = ?`
	_, err := r.db.Exec(query, command)
	return err
}

// === AUTO RESPONSES ===

func (r *SQLiteRepository) CreateAutoResponse(response *AutoResponse) error {
	query := `INSERT INTO auto_responses 
			  (keyword, response_type, sticker_path, audio_path, text_response, is_active, created_by, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	now := time.Now()
	_, err := r.db.Exec(query, response.Keyword, response.ResponseType, response.StickerPath,
		response.AudioPath, response.TextResponse, response.IsActive, response.CreatedBy, now, now)
	
	return err
}

func (r *SQLiteRepository) GetAutoResponse(keyword string) (*AutoResponse, error) {
	query := `SELECT id, keyword, response_type, sticker_path, audio_path, text_response, 
			  is_active, usage_count, created_by, created_at, updated_at 
			  FROM auto_responses WHERE keyword = ? AND is_active = 1`
	
	var response AutoResponse
	err := r.db.QueryRow(query, keyword).Scan(
		&response.ID, &response.Keyword, &response.ResponseType, &response.StickerPath,
		&response.AudioPath, &response.TextResponse, &response.IsActive, &response.UsageCount,
		&response.CreatedBy, &response.CreatedAt, &response.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &response, nil
}

func (r *SQLiteRepository) GetAutoResponsesByKeyword(text string) ([]AutoResponse, error) {
	query := `SELECT id, keyword, response_type, sticker_path, audio_path, text_response, 
			  is_active, usage_count, created_by, created_at, updated_at 
			  FROM auto_responses WHERE is_active = 1 AND ? LIKE '%' || keyword || '%'`
	
	rows, err := r.db.Query(query, text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var responses []AutoResponse
	for rows.Next() {
		var response AutoResponse
		err := rows.Scan(&response.ID, &response.Keyword, &response.ResponseType,
			&response.StickerPath, &response.AudioPath, &response.TextResponse,
			&response.IsActive, &response.UsageCount, &response.CreatedBy,
			&response.CreatedAt, &response.UpdatedAt)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	
	return responses, nil
}

func (r *SQLiteRepository) GetAllAutoResponses() ([]AutoResponse, error) {
	query := `SELECT id, keyword, response_type, sticker_path, audio_path, text_response, 
			  is_active, usage_count, created_by, created_at, updated_at 
			  FROM auto_responses ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var responses []AutoResponse
	for rows.Next() {
		var response AutoResponse
		err := rows.Scan(&response.ID, &response.Keyword, &response.ResponseType,
			&response.StickerPath, &response.AudioPath, &response.TextResponse,
			&response.IsActive, &response.UsageCount, &response.CreatedBy,
			&response.CreatedAt, &response.UpdatedAt)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	
	return responses, nil
}

func (r *SQLiteRepository) UpdateAutoResponse(response *AutoResponse) error {
	query := `UPDATE auto_responses 
			  SET response_type = ?, sticker_path = ?, audio_path = ?, text_response = ?, 
			      is_active = ?, updated_at = ? 
			  WHERE keyword = ?`
	
	_, err := r.db.Exec(query, response.ResponseType, response.StickerPath, response.AudioPath,
		response.TextResponse, response.IsActive, time.Now(), response.Keyword)
	
	return err
}

func (r *SQLiteRepository) DeleteAutoResponse(keyword string) error {
	query := `DELETE FROM auto_responses WHERE keyword = ?`
	_, err := r.db.Exec(query, keyword)
	return err
}

func (r *SQLiteRepository) IncrementAutoResponseUsage(keyword string) error {
	query := `UPDATE auto_responses SET usage_count = usage_count + 1 WHERE keyword = ?`
	_, err := r.db.Exec(query, keyword)
	return err
}

// === COMMAND USAGE LOGS ===

func (r *SQLiteRepository) LogCommandUsage(log *CommandUsageLog) error {
	query := `INSERT INTO command_usage_logs 
			  (command_type, command_value, group_jid, user_jid, response_type, success, error_message, used_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := r.db.Exec(query, log.CommandType, log.CommandValue, log.GroupJID, log.UserJID,
		log.ResponseType, log.Success, log.ErrorMessage, time.Now())
	
	return err
}

func (r *SQLiteRepository) GetCommandUsageLogs(limit int) ([]CommandUsageLog, error) {
	query := `SELECT id, command_type, command_value, group_jid, user_jid, response_type, 
			  success, error_message, used_at 
			  FROM command_usage_logs ORDER BY used_at DESC LIMIT ?`
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var logs []CommandUsageLog
	for rows.Next() {
		var log CommandUsageLog
		err := rows.Scan(&log.ID, &log.CommandType, &log.CommandValue, &log.GroupJID,
			&log.UserJID, &log.ResponseType, &log.Success, &log.ErrorMessage, &log.UsedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	
	return logs, nil
}

func (r *SQLiteRepository) GetCommandUsageStats(days int) (map[string]int, error) {
	query := `SELECT command_value, COUNT(*) as count 
			  FROM command_usage_logs 
			  WHERE used_at >= datetime('now', '-' || ? || ' days') AND success = 1
			  GROUP BY command_value ORDER BY count DESC`
	
	rows, err := r.db.Query(query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	stats := make(map[string]int)
	for rows.Next() {
		var command string
		var count int
		err := rows.Scan(&command, &count)
		if err != nil {
			return nil, err
		}
		stats[command] = count
	}
	
	return stats, nil
}

// === FORBIDDEN WORDS ===

func (r *SQLiteRepository) CreateForbiddenWord(word *ForbiddenWord) error {
	query := `INSERT INTO forbidden_words (group_jid, word, created_by, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, word.GroupJID, word.Word, word.CreatedBy, time.Now())
	return err
}

func (r *SQLiteRepository) GetForbiddenWordsByGroup(groupJID string) ([]ForbiddenWord, error) {
	query := `SELECT id, group_jid, word, created_by, created_at FROM forbidden_words WHERE group_jid = ? ORDER BY created_at DESC`
	rows, err := r.db.Query(query, groupJID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []ForbiddenWord
	for rows.Next() {
		var word ForbiddenWord
		if err := rows.Scan(&word.ID, &word.GroupJID, &word.Word, &word.CreatedBy, &word.CreatedAt); err != nil {
			return nil, err
		}
		words = append(words, word)
	}
	return words, nil
}

func (r *SQLiteRepository) DeleteForbiddenWord(id int) error {
	query := `DELETE FROM forbidden_words WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}


// === UTILITY FUNCTIONS ===

// InitializeDatabase menginisialisasi database dan menjalankan migrasi
func InitializeDatabase(dbPath string) (*sql.DB, Repository, error) {
	// Buka koneksi database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %v", err)
	}
	
	// Test koneksi
	if err := db.Ping(); err != nil {
		return nil, nil, fmt.Errorf("failed to ping database: %v", err)
	}
	
	// Jalankan migrasi
	if err := RunMigrations(db); err != nil {
		return nil, nil, fmt.Errorf("failed to run migrations: %v", err)
	}
	
	// Buat repository
	repo := NewSQLiteRepository(db)
	
	return db, repo, nil
}