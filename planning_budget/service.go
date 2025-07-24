package planning_budget

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/planning_budget/activity"
	"github.com/AMETORY/ametory-erp-modules/planning_budget/budget"
	"github.com/AMETORY/ametory-erp-modules/planning_budget/component"
	"github.com/AMETORY/ametory-erp-modules/planning_budget/kpi"
	"github.com/AMETORY/ametory-erp-modules/planning_budget/output"
	"github.com/AMETORY/ametory-erp-modules/planning_budget/strategic_objective"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type PlanningBudgetService struct {
	ctx              *context.ERPContext
	BudgetService    *budget.BudgetService
	ActivityService  *activity.ActivityService
	ComponentService *component.ComponentService
	KPIService       *kpi.KPIService
	OutputService    *output.OutputService
	StrategyService  *strategic_objective.StrategicObjectiveService
}

func NewPlanningBudgetService(ctx *context.ERPContext) *PlanningBudgetService {
	service := PlanningBudgetService{
		ctx:              ctx,
		BudgetService:    budget.NewBudgetService(ctx),
		ActivityService:  activity.NewActivityService(ctx),
		ComponentService: component.NewComponentService(ctx),
		KPIService:       kpi.NewKPIService(ctx),
		OutputService:    output.NewOutputService(ctx),
		StrategyService:  strategic_objective.NewStrategicObjectiveService(ctx),
	}

	if !ctx.SkipMigration {
		err := Migrate(ctx)
		if err != nil {
			panic(err)
		}
	}
	return &service
}

func Migrate(ctx *context.ERPContext) error {
	if ctx.SkipMigration {
		return nil
	}
	return ctx.DB.AutoMigrate(
		&models.BudgetModel{},
		&models.BudgetActivityModel{},
		&models.BudgetComponentModel{},
		&models.BudgetKPIModel{},
		&models.BudgetOutputModel{},
		&models.BudgetStrategicObjectiveModel{},
	)
}
