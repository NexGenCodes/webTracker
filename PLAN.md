# Plan: Commit and Push Changes

The goal is to commit and push the current modifications to the repository. The changes primarily involve UI enhancements, internationalization, and admin features in the frontend, as well as a new database reset script.

## Impact Analysis

### Frontend (front/)
- **UI Components**: Updates to `ShipmentForm.tsx`, `ShipmentTable.tsx`, `StatsCards.tsx`, `SuccessDisplay.tsx`, `FeatureCard.tsx`, `Header.tsx`, `Logo.tsx`.
- **Pages**: Updates to `admin/page.tsx` and the main `page.tsx`.
- **Logic & Services**: Updates to `shipment.service.ts`, `route.ts`, `logger.ts`, `utils.ts`, and types.
- **I18n**: Updates to `I18nContext.tsx`, `LanguageToggle.tsx`, and `ThemeToggle.tsx`.

### Database (database/)
- **New Script**: `database/reset.mjs` for resetting the database state.

## Breaking Changes
- None expected. The changes appear to be additive or refining existing features.

## Verification
- [ ] Run `npm run lint` and `npm run typecheck` in the `front` directory.
- [ ] Verify that the application builds successfully.
- [ ] Confirm git status is clean after commit.

## Proposed Git Commands
1. `git add .`
2. `git commit -m "Refactor UI components, enhance i18n support, and add database reset script"`
3. `git push origin main` (assuming main branch)
