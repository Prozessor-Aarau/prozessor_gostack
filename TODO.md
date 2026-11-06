# TODO

## Calendar iCal Feed

- [ ] Add public route `GET /calendar/feed.ics` — no auth, no CSRF
- [ ] Return `Content-Type: text/calendar; charset=utf-8` with all opening hours and events serialized in iCal format
- [ ] Each entry needs a stable `UID` (e.g. `opening-hour-{id}@prozessor-aarau.ch`) so calendar apps can sync updates without duplicating
- [ ] Consider `github.com/arran4/golang-ical` library or write the format directly (it's simple enough)
- [ ] Users subscribe by pasting the URL into Google Calendar → "Add calendar → From URL" — Google re-fetches periodically, no push needed

---

## Calendar Backend

### Drag-to-create (desktop only) — complex
- `mousedown` on grid → record start slot
- `mousemove` → draw ghost event preview snapped to slot grid
- `mouseup` → open pre-filled confirmation modal, actual POST happens on form submit (safer than creating on mouseup directly)

```js
let dragStart = null;

grid.addEventListener('mousedown', e => {
  dragStart = slotFromY(e.clientY);
  showGhost(dragStart, dragStart);
});

grid.addEventListener('mousemove', e => {
  if (!dragStart) return;
  showGhost(dragStart, slotFromY(e.clientY));
});

grid.addEventListener('mouseup', e => {
  const dragEnd = slotFromY(e.clientY);
  dragStart = null;
  hideGhost();
  htmx.ajax('POST', '/events/new', {
    target: '#modal',
    values: { start: slotToTime(dragStart), end: slotToTime(dragEnd) }
  });
});
```

- **Mobile:** drag-to-create doesn't work on touch — use an "add event" button that opens a time picker form instead

---

## Database — `updated_on` Trigger

The `set_updated_on()` function and `BEFORE UPDATE` trigger are installed on all existing tables via migration `20260704104933`. The trigger automatically sets `updated_on = CURRENT_TIMESTAMP` on every `UPDATE`.

**When creating a new table with an `updated_on` column**, add this to its migration:

```sql
CREATE TRIGGER trg_set_updated_on
BEFORE UPDATE ON your_new_table
FOR EACH ROW EXECUTE FUNCTION set_updated_on();
```

> Event triggers (which would attach this automatically on `CREATE TABLE`) require superuser privileges in Postgres and could not be used.

---

## Newsletter — Fallback Unsubscribe

Some recipients use mail proxies (self-hosted servers, corporate gateways) that rewrite links and mangle the UUID in the unsubscribe URL. The handler already validates the UUID and logs the corrupted ID, but those users currently see an error and cannot unsubscribe.

- [ ] Add a fallback route `GET /unsubscribe` (no `{id}`) that renders a simple form with an email input
- [ ] Add handler `POST /unsubscribe` that looks up the email in both `visitors` (set `newsletter = false`) and `newsletter_recipients` (delete row), returning the same confirmation page as the normal unsubscribe flow
- [ ] When the UUID validation fails in `HandleNewsletterUnsubscribe`, redirect to `/unsubscribe` instead of returning an error
