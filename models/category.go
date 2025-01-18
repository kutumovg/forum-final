package models

import (
	"errors"
)

// Category represents a category structure
type Category struct {
	ID              int
	Name            string
	IsControversial bool
}

// AddCategory adds a new category to the database
func AddCategory(name string) error {
	if name == "" {
		return errors.New("category name cannot be empty")
	}
	_, err := db.Exec("INSERT INTO categories (name) VALUES (?)", name)
	return err
}

// DeleteCategory deletes a category by ID
func DeleteCategory(id int) error {
	// _, err := db.Exec("DELETE FROM categories WHERE id = ?", id)
	// return err
	// Начинаем транзакцию
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Удаляем записи из post_categories
	_, err = tx.Exec("DELETE FROM post_categories WHERE category_id = ?", id)
	if err != nil {
		tx.Rollback() // Откат в случае ошибки
		return err
	}

	// Удаляем категорию из categories
	_, err = tx.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		tx.Rollback() // Откат в случае ошибки
		return err
	}

	// Фиксируем транзакцию
	return tx.Commit()
}

// UpdateCategory updates the name of a category by ID
func UpdateCategory(id int, newName string) error {
	if newName == "" {
		return errors.New("new category name cannot be empty")
	}
	_, err := db.Exec("UPDATE categories SET name = ? WHERE id = ?", newName, id)
	return err
}

// GetAllCategories retrieves all categories from the database
func GetAllCategories() ([]Category, error) {
	rows, err := db.Query("SELECT id, name, is_controversial FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.IsControversial); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func SetCategoryControversialStatus(categoryID int, isControversial bool) error {
	_, err := db.Exec("UPDATE categories SET is_controversial = ? WHERE id = ?", isControversial, categoryID)
	return err
}
