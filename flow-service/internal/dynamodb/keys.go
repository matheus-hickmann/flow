package dynamodb

import "strconv"

// PK/SK conventions for the single-table design. Port of the Java Keys.java.
// All user-scoped PKs start with "USER#{userId}#".
const (
	prefixUser = "USER#"
	prefixAcc  = "ACC#"
	prefixTxn  = "TXN#"
	prefixAuth = "AUTH#"

	SKMetadata      = "METADATA"
	SKBalanceLatest = "BALANCE#LATEST"
	SKEntryPrefix   = "ENTRY#"
	SKProfile       = "PROFILE"
	SKCategories    = "CATEGORIES"
	SKEconParams    = "ECON#PARAMS"
	SKSalary        = "SALARY"

	SKBudgetPrefix       = "BUDGET#"
	SKGoalPrefix         = "GOAL#"
	SKMerchantRulePrefix = "MERCHANT_RULE#"
	SKDebtPrefix         = "DEBT#"

	prefixGroup  = "GROUP#"
	prefixInvite = "INVITE#"

	SKMemberPrefix = "MEMBER#"
	SKInvitePrefix = "INVITE#"
)

// UserPrefix returns the per-user namespace ("USER#{id}#").
func UserPrefix(userID string) string {
	return prefixUser + userID + "#"
}

// AccountPK is the partition key for an account's metadata + balance items.
func AccountPK(userID, accountID string) string {
	return UserPrefix(userID) + prefixAcc + accountID
}

// GSI1PKAccounts groups every account belonging to a user under one GSI1 partition.
func GSI1PKAccounts(userID string) string {
	return UserPrefix(userID) + "ACCOUNT"
}

// GSI1SKAccount sorts accounts within the per-user GSI1 partition.
func GSI1SKAccount(accountID string) string {
	return prefixAcc + accountID
}

// TransactionPK is the partition key for a transaction's metadata + entries.
func TransactionPK(userID, txID string) string {
	return UserPrefix(userID) + prefixTxn + txID
}

// GSI1PKTransactions groups every transaction belonging to a user.
func GSI1PKTransactions(userID string) string {
	return UserPrefix(userID) + "TRANSACTION"
}

// GSI1SKTransaction sorts transactions within the per-user GSI1 partition.
func GSI1SKTransaction(txID string) string {
	return prefixTxn + txID
}

// EntrySK builds the sort key for one ledger entry inside a transaction.
func EntrySK(accountID string, seq int) string {
	return SKEntryPrefix + accountID + "#" + strconv.Itoa(seq)
}

// UserPK is the partition key for the user profile + categories.
func UserPK(userID string) string {
	return prefixUser + userID
}

// AuthPK is the partition key for credentials (userId + password).
func AuthPK(userID string) string {
	return prefixAuth + userID
}

// UserPlanPK is the partition key for planning items (budgets, goals, params, salary).
func UserPlanPK(userID string) string {
	return UserPrefix(userID) + "PLAN"
}

// BudgetSK builds the sort key for a budget limit.
func BudgetSK(id string) string {
	return SKBudgetPrefix + id
}

// GoalSK builds the sort key for an investment goal.
func GoalSK(id string) string {
	return SKGoalPrefix + id
}

// MerchantRuleSK builds the sort key for a merchant→category rule.
func MerchantRuleSK(merchantKey string) string {
	return SKMerchantRulePrefix + merchantKey
}

// GroupPK is the partition key for a family group and its members/invites.
func GroupPK(groupID string) string {
	return prefixGroup + groupID
}

// GSI1PKUserGroups lets us query all groups a user belongs to.
func GSI1PKUserGroups(userID string) string {
	return UserPrefix(userID) + "GROUP"
}

// GSI1SKUserGroup sorts group memberships within the per-user GSI1 partition.
func GSI1SKUserGroup(groupID string) string {
	return prefixGroup + groupID
}

// MemberSK is the sort key for a group membership item.
func MemberSK(userID string) string {
	return SKMemberPrefix + userID
}

// UserDebtPK is the partition key for all debt items belonging to a user.
func UserDebtPK(userID string) string {
	return UserPrefix(userID) + "DEBT"
}

// DebtSK builds the sort key for a single debt.
func DebtSK(id string) string {
	return SKDebtPrefix + id
}

// InvitePK is the partition key for an invite looked up by token.
func InvitePK(token string) string {
	return prefixInvite + token
}

// InviteSK is the sort key stored inside a group for listing invites.
func InviteSK(token string) string {
	return SKInvitePrefix + token
}
