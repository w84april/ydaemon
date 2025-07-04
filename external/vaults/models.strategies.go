package vaults

import (
	"github.com/yearn/ydaemon/common/bigNumber"
	"github.com/yearn/ydaemon/internal/models"
)

/**************************************************************************************************
** TExternalStrategyDetails contains detailed financial and operational data for a strategy.
**
** This structure stores key metrics about a strategy's performance and configuration in the
** Yearn ecosystem. It includes debt allocation, profit/loss tracking, fee information, and
** operational status flags that determine how the strategy interacts with its associated vault.
**
** @field TotalDebt *bigNumber.Int - The total amount of funds allocated to the strategy
** @field TotalLoss *bigNumber.Int - The cumulative losses incurred by the strategy
** @field TotalGain *bigNumber.Int - The cumulative gains generated by the strategy
** @field PerformanceFee uint64 - The fee percentage charged on strategy profits
** @field LastReport uint64 - Timestamp of the last strategy report to its vault
** @field DebtRatio uint64 - The percentage of vault funds allocated to this strategy (for v0.2.2+)
** @field InQueue bool - Whether the strategy is in the vault's withdrawal queue
**************************************************************************************************/
type TExternalStrategyDetails struct {
	TotalDebt      *bigNumber.Int `json:"totalDebt"`
	TotalLoss      *bigNumber.Int `json:"totalLoss"`
	TotalGain      *bigNumber.Int `json:"totalGain"`
	PerformanceFee uint64         `json:"performanceFee"`
	LastReport     uint64         `json:"lastReport"`
	DebtRatio      uint64         `json:"debtRatio,omitempty"` // Only > 0.2.2
	InQueue        bool           `json:"-"`
}

/**************************************************************************************************
** TStrategy represents a yield-generating strategy for Yearn vaults.
**
** This structure contains the essential information about a strategy, which is the core
** component that generates yield for Yearn vaults. Each strategy implements specific logic
** to deploy funds in various DeFi protocols to generate returns.
**
** @field Address string - The on-chain address of the strategy contract
** @field Name string - The human-readable name of the strategy
** @field Description string - A description of the strategy's approach and mechanisms
** @field Status string - The operational status of the strategy (active, not_active, unallocated)
** @field Details *TExternalStrategyDetails - Detailed performance and configuration metrics
**************************************************************************************************/
type TExternalStrategy struct {
	Address     string                    `json:"address"`
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Status      string                    `json:"status"`
	NetAPR      float64                   `json:"netAPR,omitempty"`
	Details     *TExternalStrategyDetails `json:"details,omitempty"`
}

/**************************************************************************************************
** CreateExternalStrategy creates a fully populated external strategy structure from an internal model.
**
** This function directly creates and populates a TExternalStrategy instance from a models.TStrategy.
**
** @param strategy models.TStrategy - The internal strategy model to convert
** @return TExternalStrategy - The fully populated external strategy structure
**************************************************************************************************/
func CreateExternalStrategy(strategy models.TStrategy) TExternalStrategy {
	name := strategy.DisplayName
	if name == "" {
		name = strategy.Name
	}

	// Use the Status field if it's set, otherwise determine it based on the rules
	status := string(strategy.Status)
	if status == "" {
		if strategy.IsRetired {
			status = string(models.StrategyStatusNotActive)
		} else if strategy.LastTotalDebt.IsZero() {
			status = string(models.StrategyStatusUnallocated)
		} else {
			status = string(models.StrategyStatusActive)
		}
	}

	return TExternalStrategy{
		Address:     strategy.Address.Hex(),
		Name:        name,
		Description: strategy.Description,
		Status:      status,
		NetAPR:      strategy.NetAPR,
		Details: &TExternalStrategyDetails{
			TotalDebt:      strategy.LastTotalDebt,
			TotalLoss:      strategy.LastTotalLoss,
			TotalGain:      strategy.LastTotalGain,
			PerformanceFee: strategy.LastPerformanceFee.Uint64(),
			LastReport:     strategy.LastReport.Uint64(),
			DebtRatio:      strategy.LastDebtRatio.Uint64(),
			InQueue:        strategy.IsInQueue,
		},
	}
}

/**************************************************************************************************
** ShouldBeIncluded determines whether a strategy should be included in API responses.
**
** This method evaluates whether a strategy meets the specified inclusion condition. The
** conditions control which strategies are returned in API responses based on their
** operational status and configuration:
**
** - 'all': Include all strategies regardless of status
** - 'absolute': Include only strategies with positive debt allocation
** - 'inQueue': Include only strategies in the vault's withdrawal queue
** - 'debtRatio': Include only strategies with a positive debt ratio
**
** @param condition string - The inclusion condition to check against
** @return bool - True if the strategy should be included, false otherwise
**************************************************************************************************/
func (v TExternalStrategy) ShouldBeIncluded(condition string) bool {
	if condition == `all` {
		return true
	} else if condition == `absolute` && v.Details.TotalDebt.Gt(bigNumber.Zero) {
		return true
	} else if condition == `inQueue` && v.Details.InQueue {
		return true
	} else if condition == `debtRatio` && v.Details.DebtRatio > 0 {
		return true
	}
	return false
}
