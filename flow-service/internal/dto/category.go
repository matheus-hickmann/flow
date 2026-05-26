package dto

// CategoryItem is one entry in the user's expense/income list.
type CategoryItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// CategoryList groups expense and income categories.
type CategoryList struct {
	Expense []CategoryItem `json:"expense"`
	Income  []CategoryItem `json:"income"`
}

// DefaultCategoryList mirrors the Java defaults — what we return when the user
// hasn't customized their categories yet.
func DefaultCategoryList() CategoryList {
	return CategoryList{
		Expense: []CategoryItem{
			{ID: "alimentacao", Name: "Alimentação", Color: "#f97316"},
			{ID: "moradia", Name: "Moradia", Color: "#8b5cf6"},
			{ID: "transporte", Name: "Transporte", Color: "#3b82f6"},
			{ID: "saude", Name: "Saúde", Color: "#ef4444"},
			{ID: "educacao", Name: "Educação", Color: "#eab308"},
			{ID: "lazer", Name: "Lazer", Color: "#ec4899"},
			{ID: "vestuario", Name: "Vestuário", Color: "#14b8a6"},
			{ID: "outros", Name: "Outros", Color: "#6b7280"},
		},
		Income: []CategoryItem{
			{ID: "salario", Name: "Salário", Color: "#22c55e"},
			{ID: "freelance", Name: "Freelance", Color: "#10b981"},
			{ID: "investimentos", Name: "Investimentos", Color: "#f59e0b"},
			{ID: "aluguel", Name: "Aluguel recebido", Color: "#06b6d4"},
			{ID: "outros", Name: "Outros", Color: "#6b7280"},
		},
	}
}
