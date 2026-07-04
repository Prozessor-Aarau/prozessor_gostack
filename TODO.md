# TODO

## Calendar Embed (external site)

- [ ] Create public route: `GET /calendar/{startdate}/{daycount}` — e.g. `/calendar/2026-03-27/3`
- [ ] Handler calculates `prev`/`next` startdates by subtracting/adding `daycount` days, passes pre-built URLs to templ
- [ ] Desktop renders 7-day view (`daycount=7`), mobile renders 3-day view (`daycount=3`) — same route and handler, template adjusts columns
- [ ] Add resize JS listener that re-fetches with the correct `daycount` when crossing the breakpoint — set threshold low (~`480px`) so half-screen desktop still gets the full 7-day view; on mobile viewport is fixed so this only matters for desktop edge cases
- [ ] Exempt calendar routes from CSRF in `NoSurf` middleware (same pattern as `/subscribe`)
- [ ] Return calendar fragment + swipe JS with each response (JS re-attaches via IIFE on every HTMX swap)
- [ ] Go handler pre-calculates and passes `PrevURL`/`NextURL` to templ — no date math in JS
- [ ] Templ renders them as `data-next`/`data-prev` on the `#calendar` div — swipe JS reads these and calls `htmx.ajax()`, staying completely dumb
- [ ] External site bootstraps by loading HTMX and calling `htmx.ajax('/calendar/2026-03-27/3', ...)` once to mount the calendar
- [ ] Keep all routes GET-only — no CSRF token available cross-origin

---

## Calendar iCal Feed

- [ ] Add public route `GET /calendar/feed.ics` — no auth, no CSRF
- [ ] Return `Content-Type: text/calendar; charset=utf-8` with all opening hours and events serialized in iCal format
- [ ] Each entry needs a stable `UID` (e.g. `opening-hour-{id}@prozessor-aarau.ch`) so calendar apps can sync updates without duplicating
- [ ] Consider `github.com/arran4/golang-ical` library or write the format directly (it's simple enough)
- [ ] Users subscribe by pasting the URL into Google Calendar → "Add calendar → From URL" — Google re-fetches periodically, no push needed

---

## Calendar View (frontend)

### Grid layout
- Use CSS Grid: 1 time-label column + `daycount` day columns, 1 header row + 1 row per slot
- 30min slots = 28 rows (8am–10pm) — use 15min slots (56 rows) for finer granularity (events starting at :15/:45)
- Each event spans rows via `grid-row: start / end`
- Row height responsive via CSS custom property, min height to keep slots tappable on mobile

```css
.calendar {
  --slot-height: max(20px, calc((100vh - 40px) / 28));

  display: grid;
  grid-template-columns: 60px repeat(var(--day-count), 1fr);
  grid-template-rows: 40px repeat(28, var(--slot-height));
}

.event {
  grid-column: 3;   /* wednesday */
  grid-row: 5 / 9;  /* 10:00–12:00 */
}
```

- `--day-count` set inline from templ (3 or 7 depending on viewport/request)

### Overlapping events
- Each day column is `position: relative`
- Events inside are `position: absolute`, width/left calculated from overlap group
- Pass layout values as CSS custom properties from templ:

```html
<div class="event" style="
  --start-slot: 4;
  --duration-slots: 3;
  --col: 1;
  --total-cols: 2;
">
```

```css
.event {
  position: absolute;
  top: calc(var(--start-slot) * 40px);
  height: calc(var(--duration-slots) * 40px);
  left: calc(var(--col) / var(--total-cols) * 100%);
  width: calc(100% / var(--total-cols));
}
```

> **TODO (optimization):** Current `layoutEvents` assigns columns first-come-first-served within a simple overlap group. A smarter algorithm tracks per-column occupancy and reuses freed columns, avoiding unnecessarily wide splits when earlier events end before later ones start:
> ```go
> colEnds := []time.Time{}
> for _, e := range events {
>     assigned := -1
>     for col, end := range colEnds {
>         if !e.GetStartTime().Before(end) {
>             assigned = col
>             colEnds[col] = e.GetEndTime()
>             break
>         }
>     }
>     if assigned == -1 {
>         assigned = len(colEnds)
>         colEnds = append(colEnds, e.GetEndTime())
>     }
>     e.SetColumn(assigned)
> }
> // second pass to set TotalCols per overlap group
> ```
```

### Layout math in Go (before rendering — keeps templ dumb)

Handler pre-calculates prev/next URLs — no date math in JS:

```go
func (h *Handler) Calendar(w http.ResponseWriter, r *http.Request) {
    startDate, _ := time.Parse("2006-01-02", chi.URLParam(r, "startdate"))
    dayCount, _ := strconv.Atoi(chi.URLParam(r, "daycount"))

    prev := startDate.AddDate(0, 0, -dayCount).Format("2006-01-02")
    next := startDate.AddDate(0, 0, dayCount).Format("2006-01-02")

    data.PrevURL = fmt.Sprintf("/calendar/%s/%d", prev, dayCount)
    data.NextURL = fmt.Sprintf("/calendar/%s/%d", next, dayCount)
}
```

Overlap layout calculation:

```go
type Event struct {
    Title     string
    StartTime time.Time
    EndTime   time.Time
    Column    int
    TotalCols int
}

func layoutEvents(events []Event) []Event {
    sort.Slice(events, func(i, j int) bool {
        return events[i].StartTime.Before(events[j].StartTime)
    })

    groups := [][]Event{}
    for _, e := range events {
        placed := false
        for i, g := range groups {
            for _, ge := range g {
                if e.StartTime.Before(ge.EndTime) && e.EndTime.After(ge.StartTime) {
                    groups[i] = append(groups[i], e)
                    placed = true
                    break
                }
            }
            if placed { break }
        }
        if !placed {
            groups = append(groups, []Event{e})
        }
    }

    result := []Event{}
    for _, g := range groups {
        for col, e := range g {
            e.Column = col
            e.TotalCols = len(g)
            result = append(result, e)
        }
    }
    return result
}

func timeToSlot(t time.Time, startHour int, slotMinutes int) int {
    minutes := (t.Hour()-startHour)*60 + t.Minute()
    return minutes / slotMinutes
}
```

---

## Calendar Backend

### Recurring opening hours — easy
- Bulk insert recurring rows via script or form
- No special UI needed

### Delete event — easy
- Delete button on event → `hx-delete` → handler removes row

### Supervisor signup — easy
- Click event → HTMX loads modal → POST `/events/{id}/supervisor`
- Similar pattern to existing checkin flow

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

- **Mobile:** drag-to-create doesn't work on touch — use an "add event" button that opens a time picker form instead
