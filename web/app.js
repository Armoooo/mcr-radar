(() => {
  'use strict';

  const one = (selector, root = document) => root.querySelector(selector);
  const all = (selector, root = document) => [...root.querySelectorAll(selector)];
  const state = {
    events: [],
    category: 'All',
    date: 'all',
    search: '',
    saved: new Set(readSaved()),
    mode: 'demo',
  };

  const elements = {
    grid: one('[data-events]'),
    categories: one('[data-categories]'),
    count: one('[data-results-count]'),
    mode: one('[data-mode-note]'),
    empty: one('[data-empty]'),
    search: one('[data-search]'),
    savedCount: all('[data-saved-count]'),
    savedDialog: one('[data-saved-dialog]'),
    savedList: one('[data-saved-list]'),
    template: one('#event-card-template'),
  };

  function readSaved() {
    try {
      return JSON.parse(localStorage.getItem('mcr-radar-saved') || '[]');
    } catch (_) {
      return [];
    }
  }

  function persistSaved() {
    localStorage.setItem('mcr-radar-saved', JSON.stringify([...state.saved]));
    elements.savedCount.forEach((element) => { element.textContent = state.saved.size; });
  }

  function categorySlug(value) {
    return value.toLowerCase().replace(/&/g, '').replace(/\s+/g, '-').replace(/[^a-z-]/g, '');
  }

  function eventDate(value) {
    return new Date(value);
  }

  function isSameDay(a, b) {
    return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
  }

  function isThisWeekend(date) {
    const now = new Date();
    const friday = new Date(now);
    friday.setHours(0, 0, 0, 0);
    friday.setDate(now.getDate() + ((5 - now.getDay() + 7) % 7));
    const monday = new Date(friday);
    monday.setDate(friday.getDate() + 3);
    return date >= friday && date < monday;
  }

  function filteredEvents() {
    const query = state.search.trim().toLowerCase();
    return state.events.filter((event) => {
      if (state.category !== 'All' && event.category !== state.category) return false;
      const date = eventDate(event.startsAt);
      if (state.date === 'today' && !isSameDay(date, new Date())) return false;
      if (state.date === 'weekend' && !isThisWeekend(date)) return false;
      if (query && !`${event.title} ${event.venue} ${event.area} ${event.category}`.toLowerCase().includes(query)) return false;
      return true;
    });
  }

  function renderCategories() {
    const categories = ['All', ...new Set(state.events.map((event) => event.category).filter(Boolean))];
    elements.categories.replaceChildren(...categories.map((category) => {
      const button = document.createElement('button');
      button.type = 'button';
      button.textContent = category;
      button.classList.toggle('is-active', category === state.category);
      button.setAttribute('aria-pressed', String(category === state.category));
      button.addEventListener('click', () => {
        state.category = category;
        renderCategories();
        renderEvents();
      });
      return button;
    }));
  }

  function buildCard(event) {
    const card = elements.template.content.firstElementChild.cloneNode(true);
    const date = eventDate(event.startsAt);
    card.dataset.eventId = event.id;
    card.dataset.category = categorySlug(event.category || 'Other');
    one('.event-card__category', card).textContent = event.category || 'Other';
    one('.event-card__date strong', card).textContent = date.getDate();
    one('.event-card__date span', card).textContent = date.toLocaleDateString('en-GB', { weekday: 'short', month: 'short' });
    one('h3', card).textContent = event.title;
    one('.event-card__venue', card).textContent = `${event.venue || 'Venue TBC'}${event.area ? ` · ${event.area}` : ''} · ${date.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })}`;
    const summary = one('.event-card__summary', card);
    summary.textContent = event.summary || 'Open the original listing for full event details.';
    one('.event-card__price', card).textContent = event.price || 'Check tickets';

    const image = one('img', card);
    const fallback = one('.event-card__fallback', card);
    if (event.imageUrl) {
      image.src = event.imageUrl;
      image.alt = `${event.title} event artwork`;
      image.hidden = false;
      fallback.hidden = true;
      image.addEventListener('error', () => {
        image.hidden = true;
        fallback.hidden = false;
      }, { once: true });
    }

    const sourceLink = one('.event-card__footer a', card);
    const demoLabel = one('.event-card__demo', card);
    if (event.sourceUrl) {
      sourceLink.href = event.sourceUrl;
      sourceLink.setAttribute('aria-label', `View ${event.title} on ${event.source}`);
      one('.event-card__source', sourceLink).textContent = `Via ${event.source}`;
      demoLabel.hidden = true;
    } else {
      sourceLink.hidden = true;
      demoLabel.hidden = false;
    }

    const saveButton = one('.save-button', card);
    updateSaveButton(saveButton, event);
    saveButton.addEventListener('click', () => toggleSaved(event.id));
    return card;
  }

  function updateSaveButton(button, event) {
    const saved = state.saved.has(event.id);
    button.classList.toggle('is-saved', saved);
    button.setAttribute('aria-label', `${saved ? 'Remove' : 'Save'} ${event.title}`);
    button.setAttribute('aria-pressed', String(saved));
  }

  function renderEvents() {
    const events = filteredEvents();
    elements.grid.replaceChildren(...events.map(buildCard));
    elements.grid.setAttribute('aria-busy', 'false');
    elements.count.textContent = `${events.length} ${events.length === 1 ? 'event' : 'events'} found`;
    elements.empty.hidden = events.length > 0;
    elements.grid.hidden = events.length === 0;
  }

  function toggleSaved(id) {
    if (state.saved.has(id)) state.saved.delete(id);
    else state.saved.add(id);
    persistSaved();
    renderEvents();
    if (elements.savedDialog.open) renderSaved();
  }

  function renderSaved() {
    const events = state.events.filter((event) => state.saved.has(event.id));
    if (!events.length) {
      const message = document.createElement('p');
      message.className = 'saved-empty';
      message.textContent = 'Nothing saved yet. Tap the bookmark on an event to build your shortlist.';
      elements.savedList.replaceChildren(message);
      return;
    }
    elements.savedList.replaceChildren(...events.map((event) => {
      const date = eventDate(event.startsAt);
      const item = document.createElement('article');
      item.className = 'saved-item';
      const dateBlock = document.createElement('div');
      dateBlock.className = 'saved-item__date';
      const day = document.createElement('strong');
      day.textContent = date.getDate();
      const month = document.createElement('span');
      month.textContent = date.toLocaleDateString('en-GB', { month: 'short' });
      dateBlock.append(day, month);
      const copy = document.createElement('div');
      const title = document.createElement('h3');
      title.textContent = event.title;
      const venue = document.createElement('p');
      venue.textContent = `${event.venue} · ${date.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })}`;
      copy.append(title, venue);
      const remove = document.createElement('button');
      remove.type = 'button';
      remove.textContent = 'Remove';
      remove.setAttribute('aria-label', `Remove ${event.title} from saved events`);
      remove.addEventListener('click', () => toggleSaved(event.id));
      item.append(dateBlock, copy, remove);
      return item;
    }));
  }

  function clearFilters() {
    state.category = 'All';
    state.date = 'all';
    state.search = '';
    elements.search.value = '';
    all('[data-date]').forEach((button) => button.classList.toggle('is-active', button.dataset.date === 'all'));
    renderCategories();
    renderEvents();
  }

  function pickEvent() {
    const events = filteredEvents();
    if (!events.length) {
      clearFilters();
      return;
    }
    const event = events[Math.floor(Math.random() * events.length)];
    const card = one(`[data-event-id="${CSS.escape(event.id)}"]`);
    if (!card) return;
    card.scrollIntoView({ behavior: 'smooth', block: 'center' });
    card.classList.remove('is-picked');
    requestAnimationFrame(() => card.classList.add('is-picked'));
    window.setTimeout(() => card.classList.remove('is-picked'), 1800);
  }

  function weatherIcon(code) {
    if (code >= 95) return 'ϟ';
    if (code >= 71) return '❄';
    if (code >= 51) return '☂';
    if (code >= 2) return '◒';
    return '☀';
  }

  async function loadWeather() {
    try {
      const response = await fetch('/api/weather');
      if (!response.ok) throw new Error('weather unavailable');
      const payload = await response.json();
      const days = payload.forecasts.slice(0, 3);
      one('[data-weather-summary]').textContent = days[0]?.recommendation || 'Forecast ready';
      one('[data-weather-days]').replaceChildren(...days.map((forecast) => {
        const day = document.createElement('div');
        day.className = 'weather-day';
        const icon = document.createElement('span');
        icon.className = 'weather-day__icon';
        icon.setAttribute('aria-hidden', 'true');
        icon.textContent = weatherIcon(forecast.weatherCode);
        const label = document.createElement('span');
        label.textContent = new Date(`${forecast.date}T12:00:00`).toLocaleDateString('en-GB', { weekday: 'short' });
        const temperature = document.createElement('strong');
        temperature.textContent = `${Math.round(forecast.maximum)}° / ${Math.round(forecast.minimum)}°`;
        day.append(icon, label, temperature);
        return day;
      }));
      one('[data-weather-credit]').hidden = false;
    } catch (_) {
      one('[data-weather-summary]').textContent = 'Manchester weather: unpredictable';
      one('[data-weather-days]').replaceChildren();
    }
  }

  async function loadEvents() {
    try {
      const response = await fetch('/api/events');
      if (!response.ok) throw new Error('events unavailable');
      const payload = await response.json();
      state.events = payload.events || [];
      state.mode = payload.mode;
      elements.mode.textContent = payload.mode === 'demo' ? 'Preview data · add free API keys for live events' : 'Live event data';
      renderCategories();
      renderEvents();
    } catch (_) {
      elements.count.textContent = 'The radar is temporarily offline';
      elements.mode.textContent = 'Please try again shortly';
      elements.grid.setAttribute('aria-busy', 'false');
    }
  }

  elements.search.addEventListener('input', (event) => {
    state.search = event.target.value;
    renderEvents();
  });
  all('[data-date]').forEach((button) => button.addEventListener('click', () => {
    state.date = button.dataset.date;
    all('[data-date]').forEach((item) => {
      const active = item === button;
      item.classList.toggle('is-active', active);
      item.setAttribute('aria-pressed', String(active));
    });
    renderEvents();
  }));
  one('[data-clear-filters]').addEventListener('click', clearFilters);
  one('[data-pick-event]').addEventListener('click', pickEvent);
  one('[data-show-saved]').addEventListener('click', () => {
    renderSaved();
    elements.savedDialog.showModal();
  });
  one('[data-close-saved]').addEventListener('click', () => elements.savedDialog.close());
  elements.savedDialog.addEventListener('click', (event) => {
    if (event.target === elements.savedDialog) elements.savedDialog.close();
  });

  persistSaved();
  Promise.allSettled([loadEvents(), loadWeather()]);
})();
