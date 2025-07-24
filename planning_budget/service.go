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

// NewPlanningBudgetService creates a new instance of PlanningBudgetService.
//
// It takes an ERPContext as parameter and returns a pointer to a PlanningBudgetService.
//
// It uses the ERPContext to initialize the BudgetService, ActivityService, ComponentService, KPIService,
// OutputService, and StrategicObjectiveService.
// Additionally, it calls the Migrate method of the PlanningBudgetService to create the necessary database schema,
// unless SkipMigration is set to true in the ERPContext.
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

// Migrate migrates the database schema for the PlanningBudgetService.
//
// If the SkipMigration flag is set to true in the context, this method
// will not perform any migration and will return nil. Otherwise, it will
// attempt to auto-migrate the database to include the BudgetModel,
// BudgetActivityModel, BudgetComponentModel, BudgetKPIModel,
// BudgetOutputModel, and BudgetStrategicObjectiveModel schemas.
// If the migration process encounters an error, it will return that error.
// Otherwise, it will return nil upon successful migration.
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
