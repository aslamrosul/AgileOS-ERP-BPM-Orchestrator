package handlers

import (
	"net/http"
	"time"

	"agileos-backend/logger"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ============================================
// FINANCIAL REPORTS HANDLERS
// ============================================

// GetBalanceSheet generates balance sheet report
func (h *AccountingHandler) GetBalanceSheet(c *gin.Context) {
	asOfDateStr := c.Query("as_of_date")
	if asOfDateStr == "" {
		asOfDateStr = time.Now().Format("2006-01-02")
	}

	asOfDate, err := time.Parse("2006-01-02", asOfDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// Get all accounts
	accounts, err := h.db.QuerySlice(
		"SELECT * FROM account WHERE is_active = true ORDER BY account_code ASC",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get accounts", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate balance sheet"})
		return
	}

	// Initialize report
	report := models.BalanceSheetReport{
		AsOfDate:         asOfDate,
		TotalAssets:      decimal.Zero,
		TotalLiabilities: decimal.Zero,
		TotalEquity:      decimal.Zero,
	}

	var assets, liabilities, equity []models.AccountBalance

	// Calculate balances for each account
	for _, acc := range accounts {
		account := acc.(map[string]interface{})
		accountType := account["account_type"].(string)
		accountCode := account["account_code"].(string)
		accountName := account["account_name"].(string)

		// Get account balance from journal entries
		balance := h.calculateAccountBalance(accountCode, asOfDate)

		accountBalance := models.AccountBalance{
			AccountCode: accountCode,
			AccountName: accountName,
			Balance:     balance,
		}

		switch accountType {
		case "asset":
			assets = append(assets, accountBalance)
			report.TotalAssets = report.TotalAssets.Add(balance)
		case "liability":
			liabilities = append(liabilities, accountBalance)
			report.TotalLiabilities = report.TotalLiabilities.Add(balance)
		case "equity":
			equity = append(equity, accountBalance)
			report.TotalEquity = report.TotalEquity.Add(balance)
		}
	}

	report.Assets = models.BalanceSheetSection{
		Accounts: assets,
		Total:    report.TotalAssets,
	}
	report.Liabilities = models.BalanceSheetSection{
		Accounts: liabilities,
		Total:    report.TotalLiabilities,
	}
	report.Equity = models.BalanceSheetSection{
		Accounts: equity,
		Total:    report.TotalEquity,
	}

	c.JSON(http.StatusOK, report)
}

// GetProfitLoss generates profit & loss report
func (h *AccountingHandler) GetProfitLoss(c *gin.Context) {
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	if fromDateStr == "" || toDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from_date and to_date are required"})
		return
	}

	fromDate, err := time.Parse("2006-01-02", fromDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format"})
		return
	}

	toDate, err := time.Parse("2006-01-02", toDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format"})
		return
	}

	// Get revenue and expense accounts
	accounts, err := h.db.QuerySlice(
		"SELECT * FROM account WHERE (account_type = 'revenue' OR account_type = 'expense') AND is_active = true ORDER BY account_code ASC",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get accounts", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate profit & loss"})
		return
	}

	report := models.ProfitLossReport{
		FromDate:      fromDate,
		ToDate:        toDate,
		TotalRevenue:  decimal.Zero,
		TotalExpenses: decimal.Zero,
	}

	var revenue, expenses []models.AccountBalance

	for _, acc := range accounts {
		account := acc.(map[string]interface{})
		accountType := account["account_type"].(string)
		accountCode := account["account_code"].(string)
		accountName := account["account_name"].(string)

		balance := h.calculateAccountBalanceForPeriod(accountCode, fromDate, toDate)

		accountBalance := models.AccountBalance{
			AccountCode: accountCode,
			AccountName: accountName,
			Balance:     balance,
		}

		if accountType == "revenue" {
			revenue = append(revenue, accountBalance)
			report.TotalRevenue = report.TotalRevenue.Add(balance)
		} else if accountType == "expense" {
			expenses = append(expenses, accountBalance)
			report.TotalExpenses = report.TotalExpenses.Add(balance)
		}
	}

	report.Revenue = revenue
	report.Expenses = expenses
	report.GrossProfit = report.TotalRevenue.Sub(report.TotalExpenses)
	report.NetProfit = report.GrossProfit

	c.JSON(http.StatusOK, report)
}

// GetCashFlow generates cash flow statement
func (h *AccountingHandler) GetCashFlow(c *gin.Context) {
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	if fromDateStr == "" || toDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from_date and to_date are required"})
		return
	}

	fromDate, err := time.Parse("2006-01-02", fromDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format"})
		return
	}

	toDate, err := time.Parse("2006-01-02", toDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format"})
		return
	}

	report := models.CashFlowReport{
		FromDate:             fromDate,
		ToDate:               toDate,
		NetCashFromOperating: decimal.Zero,
		NetCashFromInvesting: decimal.Zero,
		NetCashFromFinancing: decimal.Zero,
		NetIncreaseInCash:    decimal.Zero,
		CashAtBeginning:      decimal.Zero,
		CashAtEnd:            decimal.Zero,
	}

	// Simplified cash flow calculation
	// In production, this would analyze journal entries by category
	report.OperatingActivities = []models.CashFlowItem{
		{Description: "Net Income", Amount: decimal.NewFromInt(1000000)},
		{Description: "Depreciation", Amount: decimal.NewFromInt(50000)},
	}
	report.NetCashFromOperating = decimal.NewFromInt(1050000)

	report.InvestingActivities = []models.CashFlowItem{
		{Description: "Purchase of Equipment", Amount: decimal.NewFromInt(-200000)},
	}
	report.NetCashFromInvesting = decimal.NewFromInt(-200000)

	report.FinancingActivities = []models.CashFlowItem{
		{Description: "Loan Proceeds", Amount: decimal.NewFromInt(500000)},
	}
	report.NetCashFromFinancing = decimal.NewFromInt(500000)

	report.NetIncreaseInCash = report.NetCashFromOperating.
		Add(report.NetCashFromInvesting).
		Add(report.NetCashFromFinancing)

	c.JSON(http.StatusOK, report)
}

// GetTrialBalance generates trial balance report
func (h *AccountingHandler) GetTrialBalance(c *gin.Context) {
	asOfDateStr := c.Query("as_of_date")
	if asOfDateStr == "" {
		asOfDateStr = time.Now().Format("2006-01-02")
	}

	asOfDate, err := time.Parse("2006-01-02", asOfDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	accounts, err := h.db.QuerySlice(
		"SELECT * FROM account WHERE is_active = true ORDER BY account_code ASC",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get accounts", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate trial balance"})
		return
	}

	report := models.TrialBalanceReport{
		AsOfDate:    asOfDate,
		TotalDebit:  decimal.Zero,
		TotalCredit: decimal.Zero,
	}

	var trialBalanceAccounts []models.TrialBalanceAccount

	for _, acc := range accounts {
		account := acc.(map[string]interface{})
		accountCode := account["account_code"].(string)
		accountName := account["account_name"].(string)
		accountType := account["account_type"].(string)

		balance := h.calculateAccountBalance(accountCode, asOfDate)

		var debit, credit decimal.Decimal
		if balance.GreaterThan(decimal.Zero) {
			debit = balance
			report.TotalDebit = report.TotalDebit.Add(debit)
		} else {
			credit = balance.Abs()
			report.TotalCredit = report.TotalCredit.Add(credit)
		}

		trialBalanceAccounts = append(trialBalanceAccounts, models.TrialBalanceAccount{
			AccountCode: accountCode,
			AccountName: accountName,
			AccountType: models.AccountType(accountType),
			Debit:       debit,
			Credit:      credit,
		})
	}

	report.Accounts = trialBalanceAccounts
	report.IsBalanced = report.TotalDebit.Equal(report.TotalCredit)

	c.JSON(http.StatusOK, report)
}

// GetGeneralLedger generates general ledger report
func (h *AccountingHandler) GetGeneralLedger(c *gin.Context) {
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")
	accountID := c.Query("account_id")

	if fromDateStr == "" || toDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from_date and to_date are required"})
		return
	}

	fromDate, err := time.Parse("2006-01-02", fromDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format"})
		return
	}

	toDate, err := time.Parse("2006-01-02", toDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format"})
		return
	}

	// Get journal entries for the period
	query := `SELECT * FROM journal_line 
		WHERE entry_date >= $from_date AND entry_date <= $to_date`
	
	params := map[string]interface{}{
		"from_date": fromDate,
		"to_date":   toDate,
	}

	if accountID != "" {
		query += " AND account_id = $account_id"
		params["account_id"] = accountID
	}

	query += " ORDER BY entry_date ASC"

	entries, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get journal entries", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate general ledger"})
		return
	}

	report := models.GeneralLedgerReport{
		FromDate:       fromDate,
		ToDate:         toDate,
		OpeningBalance: decimal.Zero,
		TotalDebit:     decimal.Zero,
		TotalCredit:    decimal.Zero,
		ClosingBalance: decimal.Zero,
	}

	var glEntries []models.GeneralLedgerEntry
	runningBalance := report.OpeningBalance

	for _, entry := range entries {
		e := entry.(map[string]interface{})
		
		debit := decimal.Zero
		credit := decimal.Zero
		
		if e["debit"] != nil {
			debit = decimal.NewFromFloat(e["debit"].(float64))
		}
		if e["credit"] != nil {
			credit = decimal.NewFromFloat(e["credit"].(float64))
		}

		runningBalance = runningBalance.Add(debit).Sub(credit)
		report.TotalDebit = report.TotalDebit.Add(debit)
		report.TotalCredit = report.TotalCredit.Add(credit)

		glEntries = append(glEntries, models.GeneralLedgerEntry{
			Date:          e["entry_date"].(time.Time),
			JournalNumber: e["journal_number"].(string),
			Description:   e["description"].(string),
			Debit:         debit,
			Credit:        credit,
			Balance:       runningBalance,
		})
	}

	report.Entries = glEntries
	report.ClosingBalance = runningBalance

	c.JSON(http.StatusOK, report)
}

// Helper functions

func (h *AccountingHandler) calculateAccountBalance(accountCode string, asOfDate time.Time) decimal.Decimal {
	// Query journal lines for this account up to the date
	lines, err := h.db.QuerySlice(
		`SELECT debit, credit FROM journal_line 
		WHERE account_code = $account_code AND entry_date <= $as_of_date`,
		map[string]interface{}{
			"account_code": accountCode,
			"as_of_date":   asOfDate,
		},
	)
	if err != nil {
		return decimal.Zero
	}

	balance := decimal.Zero
	for _, line := range lines {
		l := line.(map[string]interface{})
		if l["debit"] != nil {
			balance = balance.Add(decimal.NewFromFloat(l["debit"].(float64)))
		}
		if l["credit"] != nil {
			balance = balance.Sub(decimal.NewFromFloat(l["credit"].(float64)))
		}
	}

	return balance
}

func (h *AccountingHandler) calculateAccountBalanceForPeriod(accountCode string, fromDate, toDate time.Time) decimal.Decimal {
	lines, err := h.db.QuerySlice(
		`SELECT debit, credit FROM journal_line 
		WHERE account_code = $account_code 
		AND entry_date >= $from_date 
		AND entry_date <= $to_date`,
		map[string]interface{}{
			"account_code": accountCode,
			"from_date":    fromDate,
			"to_date":      toDate,
		},
	)
	if err != nil {
		return decimal.Zero
	}

	balance := decimal.Zero
	for _, line := range lines {
		l := line.(map[string]interface{})
		if l["debit"] != nil {
			balance = balance.Add(decimal.NewFromFloat(l["debit"].(float64)))
		}
		if l["credit"] != nil {
			balance = balance.Sub(decimal.NewFromFloat(l["credit"].(float64)))
		}
	}

	return balance
}
