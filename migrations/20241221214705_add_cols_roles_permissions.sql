-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles
ALTER COLUMN role_name SET NOT NULL,
ALTER COLUMN created_on SET NOT NULL,
ALTER COLUMN updated_on SET NOT NULL,
ADD COLUMN view_visitors BOOL DEFAULT false NOT NULL,
ADD COLUMN edit_visitor BOOL DEFAULT false NOT NULL,
ADD COLUMN edit_visitor_membership BOOL DEFAULT false NOT NULL,
ADD COLUMN edit_visitor_role BOOL DEFAULT false NOT NULL,
ADD COLUMN view_machines BOOL DEFAULT false NOT NULL,
ADD COLUMN view_badges BOOL DEFAULT false NOT NULL,
ADD COLUMN assign_badge BOOL DEFAULT false NOT NULL,
ADD COLUMN edit_badge BOOL DEFAULT false NOT NULL,
ADD COLUMN view_abonnements BOOL DEFAULT false NOT NULL,
ADD COLUMN add_abonnement BOOL DEFAULT false NOT NULL,
ADD COLUMN edit_abonnement BOOL DEFAULT false NOT NULL,
ADD COLUMN view_reminders BOOL DEFAULT false NOT NULL,
ADD COLUMN add_reminder BOOL DEFAULT false NOT NULL,
ADD COLUMN view_newsletters BOOL DEFAULT false NOT NULL,
ADD COLUMN add_newsletter BOOL DEFAULT false NOT NULL,
ADD COLUMN view_logistics BOOL DEFAULT false NOT NULL,
ADD COLUMN view_statistics BOOL DEFAULT false NOT NULL,
ADD COLUMN view_logs BOOL DEFAULT false NOT NULL,
ADD COLUMN view_finances BOOL DEFAULT false NOT NULL,
ADD COLUMN add_finances BOOL DEFAULT false NOT NULL,
ADD COLUMN edit_finances BOOL DEFAULT false NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE roles
ALTER COLUMN role_name DROP NOT NULL,
ALTER COLUMN created_on DROP NOT NULL,
ALTER COLUMN updated_on DROP NOT NULL,
DROP COLUMN view_visitors,
DROP COLUMN edit_visitor,
DROP COLUMN edit_visitor_membership,
DROP COLUMN edit_visitor_role,
DROP COLUMN view_machines,
DROP COLUMN view_badges,
DROP COLUMN assign_badge,
DROP COLUMN edit_badge,
DROP COLUMN view_abonnements,
DROP COLUMN add_abonnement,
DROP COLUMN edit_abonnement,
DROP COLUMN view_reminders,
DROP COLUMN add_reminder,
DROP COLUMN view_newsletters,
DROP COLUMN add_newsletter,
DROP COLUMN view_logistics,
DROP COLUMN view_statistics,
DROP COLUMN view_logs,
DROP COLUMN view_finances,
DROP COLUMN add_finances,
DROP COLUMN edit_finances;
-- +goose StatementEnd
