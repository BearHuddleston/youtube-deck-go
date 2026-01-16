/**
 * YouTube Deck Frontend Application
 * Enhanced user interactions and accessibility features
 */

(function() {
  'use strict';

  // ===== THEME MANAGEMENT =====
  const ThemeManager = {
    STORAGE_KEY: 'youtube-deck-theme',
    DARK: 'dark',
    LIGHT: 'light',

    init() {
      const savedTheme = localStorage.getItem(this.STORAGE_KEY);
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      const theme = savedTheme || (prefersDark ? this.DARK : this.DARK); // Default to dark

      this.setTheme(theme, false);
      this.bindEvents();
    },

    setTheme(theme, save = true) {
      document.documentElement.setAttribute('data-theme', theme);
      if (save) {
        localStorage.setItem(this.STORAGE_KEY, theme);
      }
      this.updateToggleButton(theme);
    },

    toggle() {
      const current = document.documentElement.getAttribute('data-theme') || this.DARK;
      const newTheme = current === this.DARK ? this.LIGHT : this.DARK;
      this.setTheme(newTheme);
    },

    updateToggleButton(theme) {
      const buttons = document.querySelectorAll('.theme-toggle');
      buttons.forEach(btn => {
        const sunIcon = btn.querySelector('.theme-toggle__icon--sun');
        const moonIcon = btn.querySelector('.theme-toggle__icon--moon');
        if (sunIcon && moonIcon) {
          sunIcon.style.display = theme === this.DARK ? 'block' : 'none';
          moonIcon.style.display = theme === this.LIGHT ? 'block' : 'none';
        }
      });
    },

    bindEvents() {
      // Listen for system theme changes
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        if (!localStorage.getItem(this.STORAGE_KEY)) {
          this.setTheme(e.matches ? this.DARK : this.LIGHT, false);
        }
      });

      // Bind toggle buttons
      document.addEventListener('click', (e) => {
        const toggle = e.target.closest('.theme-toggle');
        if (toggle) {
          this.toggle();
        }
      });
    }
  };

  // ===== TOAST NOTIFICATIONS =====
  const ToastManager = {
    queue: [],
    isShowing: false,
    duration: 4000,

    show(message, type = 'success') {
      this.queue.push({ message, type });
      if (!this.isShowing) {
        this.processQueue();
      }
    },

    processQueue() {
      if (this.queue.length === 0) {
        this.isShowing = false;
        return;
      }

      this.isShowing = true;
      const { message, type } = this.queue.shift();

      const toast = document.getElementById('toast');
      const messageEl = document.getElementById('toast-message');

      if (!toast || !messageEl) return;

      // Update message and type
      messageEl.textContent = message;
      toast.className = toast.className.replace(/toast--(success|error|warning)/g, '');
      toast.classList.add(`toast--${type}`);

      // Update icon based on type
      const icon = toast.querySelector('svg');
      if (icon) {
        const iconPaths = {
          success: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>',
          error: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>',
          warning: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>'
        };
        const colors = { success: '#10b981', error: '#ef4444', warning: '#f59e0b' };
        icon.innerHTML = iconPaths[type] || iconPaths.success;
        icon.style.color = colors[type] || colors.success;
      }

      // Show toast
      toast.classList.remove('translate-y-20', 'opacity-0');
      toast.classList.add('toast--visible');

      // Hide after duration
      setTimeout(() => {
        toast.classList.add('translate-y-20', 'opacity-0');
        toast.classList.remove('toast--visible');

        setTimeout(() => {
          this.processQueue();
        }, 300);
      }, this.duration);
    }
  };

  // ===== KEYBOARD NAVIGATION =====
  const KeyboardNav = {
    init() {
      document.addEventListener('keydown', this.handleKeyDown.bind(this));
    },

    handleKeyDown(e) {
      // Close modals with Escape
      if (e.key === 'Escape') {
        this.closeModal();
      }

      // Focus search with Cmd/Ctrl + K
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        this.openSearch();
      }

      // Navigate with arrow keys when focused on list items
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
        this.navigateList(e);
      }
    },

    closeModal() {
      const modal = document.getElementById('modal');
      if (modal && modal.innerHTML.trim()) {
        modal.innerHTML = '';
      }
    },

    openSearch() {
      const searchBtn = document.querySelector('[hx-get="/search"]');
      if (searchBtn) {
        searchBtn.click();
      }
    },

    navigateList(e) {
      const focusable = document.activeElement;
      const list = focusable?.closest('[role="listbox"], [role="list"]');

      if (!list) return;

      const items = Array.from(list.querySelectorAll('[role="option"], [role="listitem"]'));
      const currentIndex = items.indexOf(focusable);

      if (currentIndex === -1) return;

      e.preventDefault();

      let nextIndex;
      if (e.key === 'ArrowDown') {
        nextIndex = Math.min(currentIndex + 1, items.length - 1);
      } else {
        nextIndex = Math.max(currentIndex - 1, 0);
      }

      items[nextIndex]?.focus();
    }
  };

  // ===== INTERSECTION OBSERVER FOR ANIMATIONS =====
  const AnimationObserver = {
    init() {
      if (!('IntersectionObserver' in window)) return;

      const options = {
        root: null,
        rootMargin: '0px',
        threshold: 0.1
      };

      this.observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            entry.target.classList.add('animate-fade-in-up');
            this.observer.unobserve(entry.target);
          }
        });
      }, options);

      this.observe();
    },

    observe() {
      document.querySelectorAll('.animate-on-scroll').forEach(el => {
        this.observer.observe(el);
      });
    }
  };

  // ===== CSRF TOKEN HANDLER =====
  const CSRFHandler = {
    init() {
      document.body.addEventListener('htmx:configRequest', (evt) => {
        const match = document.cookie.match(/csrf_token=([^;]+)/);
        if (match) {
          evt.detail.headers['X-CSRF-Token'] = match[1];
        }
      });
    }
  };

  // ===== SORTABLE INITIALIZATION =====
  const SortableManager = {
    init() {
      this.initSidebar();
      this.initDeck();
      this.bindHTMXEvents();
    },

    getCsrfToken() {
      const match = document.cookie.match(/csrf_token=([^;]+)/);
      return match ? match[1] : '';
    },

    initSidebar() {
      const sidebarList = document.getElementById('sidebar-list');
      if (sidebarList && !sidebarList._sortable && typeof Sortable !== 'undefined') {
        sidebarList._sortable = new Sortable(sidebarList, {
          animation: 150,
          ghostClass: 'sortable-ghost',
          chosenClass: 'sortable-chosen',
          dragClass: 'sortable-drag',
          onEnd: (evt) => {
            const ids = [...evt.to.children]
              .filter(el => el.dataset.id)
              .map(el => el.dataset.id);

            fetch('/subscriptions/reorder', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': this.getCsrfToken()
              },
              body: JSON.stringify({ ids, context: 'sidebar' })
            }).catch(err => console.error('Reorder failed:', err));
          }
        });
      }
    },

    initDeck() {
      const deckColumns = document.getElementById('deck-columns');
      if (deckColumns && !deckColumns._sortable && typeof Sortable !== 'undefined') {
        deckColumns._sortable = new Sortable(deckColumns, {
          animation: 150,
          handle: '.column-handle',
          ghostClass: 'sortable-ghost',
          chosenClass: 'sortable-chosen',
          dragClass: 'sortable-drag',
          onEnd: (evt) => {
            const ids = [...evt.to.children]
              .filter(el => el.dataset.id)
              .map(el => el.dataset.id);

            fetch('/subscriptions/reorder', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': this.getCsrfToken()
              },
              body: JSON.stringify({ ids, context: 'columns' })
            }).catch(err => console.error('Reorder failed:', err));
          }
        });
      }
    },

    bindHTMXEvents() {
      document.body.addEventListener('htmx:afterSwap', (evt) => {
        if (evt.detail.target.id === 'deck-columns' || evt.detail.target.id === 'sidebar-list') {
          setTimeout(() => {
            this.initSidebar();
            this.initDeck();
          }, 0);
        }
      });

      // Re-init on page navigation
      document.body.addEventListener('htmx:afterSettle', () => {
        setTimeout(() => {
          this.initSidebar();
          this.initDeck();
        }, 0);
      });
    }
  };

  // ===== ACTIVE COLUMNS SECTION MANAGER =====
  const ActiveColumnsManager = {
    update() {
      const section = document.getElementById('active-columns-section');
      const list = document.getElementById('active-columns-list');
      const countEl = document.getElementById('active-columns-count');

      if (!section || !list) return;

      const chips = list.querySelectorAll('[id^="active-chip-"]');

      if (chips.length === 0) {
        section.classList.add('hidden');
        section.classList.remove('border-b', 'border-zinc-800');
      } else {
        section.classList.remove('hidden');
        section.classList.add('border-b', 'border-zinc-800');
        if (countEl) countEl.textContent = chips.length;
      }
    }
  };

  // ===== LOADING STATE MANAGER =====
  const LoadingManager = {
    showSkeleton(container, count = 3) {
      const skeletons = [];
      for (let i = 0; i < count; i++) {
        skeletons.push(`
          <div class="skeleton-card animate-pulse">
            <div class="skeleton skeleton-thumbnail mb-3"></div>
            <div class="skeleton skeleton-text"></div>
            <div class="skeleton skeleton-text" style="width: 60%"></div>
          </div>
        `);
      }
      container.innerHTML = skeletons.join('');
    }
  };

  // ===== IMAGE LAZY LOADING =====
  const LazyLoadManager = {
    init() {
      if ('loading' in HTMLImageElement.prototype) {
        // Native lazy loading supported
        document.querySelectorAll('img[data-src]').forEach(img => {
          img.src = img.dataset.src;
        });
      } else {
        // Fallback for older browsers
        this.initObserver();
      }
    },

    initObserver() {
      if (!('IntersectionObserver' in window)) return;

      const imageObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            const img = entry.target;
            img.src = img.dataset.src;
            img.removeAttribute('data-src');
            imageObserver.unobserve(img);
          }
        });
      });

      document.querySelectorAll('img[data-src]').forEach(img => {
        imageObserver.observe(img);
      });
    }
  };

  // ===== FOCUS TRAP FOR MODALS =====
  const FocusTrap = {
    trap(element) {
      const focusableElements = element.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      const firstElement = focusableElements[0];
      const lastElement = focusableElements[focusableElements.length - 1];

      const handleTabKey = (e) => {
        if (e.key !== 'Tab') return;

        if (e.shiftKey) {
          if (document.activeElement === firstElement) {
            lastElement.focus();
            e.preventDefault();
          }
        } else {
          if (document.activeElement === lastElement) {
            firstElement.focus();
            e.preventDefault();
          }
        }
      };

      element.addEventListener('keydown', handleTabKey);
      firstElement?.focus();

      return () => {
        element.removeEventListener('keydown', handleTabKey);
      };
    }
  };

  // ===== HTMX EVENT HANDLERS =====
  const HTMXEventHandlers = {
    init() {
      // Show toast on custom events
      document.body.addEventListener('showToast', (e) => {
        ToastManager.show(e.detail.value, e.detail.type || 'success');
      });

      // Handle modal focus trap
      document.body.addEventListener('htmx:afterSwap', (evt) => {
        if (evt.detail.target.id === 'modal') {
          const modal = document.querySelector('.modal, #search-modal');
          if (modal) {
            FocusTrap.trap(modal);
          }
        }
      });

      // Add animations to newly loaded content
      document.body.addEventListener('htmx:afterSettle', () => {
        AnimationObserver.observe?.();
        LazyLoadManager.init();
      });
    }
  };

  // ===== INITIALIZE APPLICATION =====
  function init() {
    ThemeManager.init();
    KeyboardNav.init();
    CSRFHandler.init();
    AnimationObserver.init();
    LazyLoadManager.init();
    HTMXEventHandlers.init();

    // Initialize sortables after DOM is ready
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', () => {
        SortableManager.init();
      });
    } else {
      SortableManager.init();
    }
  }

  // Expose utilities globally
  window.YTDeck = {
    ThemeManager,
    ToastManager,
    LoadingManager,
    ActiveColumnsManager,
    FocusTrap
  };

  // Initialize on load
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

})();
