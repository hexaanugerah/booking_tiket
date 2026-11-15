import { create } from "zustand";
import { createJSONStorage, persist, type StateStorage } from "zustand/middleware";
import {
  AUTH_COOKIE_NAME,
  AUTH_STORAGE_KEY,
  SESSION_DURATION_MS,
  type AppRole,
} from "@/config/auth.config";

// Auth store — sesi login dipersist ke localStorage + cookie ringan (cookie dipakai middleware edge untuk redirect cepat sebelum hydrate). Bukan keamanan produksi.

export interface SessionUser {
  id: string;
  name: string;
  email: string;
  phone: string;
  role: AppRole;
  gate?: string;
  shift?: string;
}

interface AuthState {
  user: SessionUser | null;
  expiresAt: number | null;
  /** "remember me" — sesi lebih lama & bertahan setelah browser ditutup. */
  remember: boolean;
  login: (user: SessionUser, remember: boolean) => void;
  updateUser: (patch: Partial<Pick<SessionUser, "name" | "email" | "phone">>) => void;
  logout: () => void;
  /** Dipanggil provider: hapus user jika sesi kedaluwarsa. */
  checkExpiry: () => void;
}

function setCookie(name: string, value: string, maxAgeSeconds: number) {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${maxAgeSeconds}; samesite=lax`;
}

function clearCookie(name: string) {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=; path=/; max-age=0; samesite=lax`;
}

// A blocked/full browser storage must not turn a valid login response into a
// failed login. The session still works in memory for the current tab.
const safeStorage: StateStorage = {
  getItem: (name) => {
    try {
      return typeof window === "undefined" ? null : window.localStorage.getItem(name);
    } catch {
      return null;
    }
  },
  setItem: (name, value) => {
    try {
      if (typeof window !== "undefined") window.localStorage.setItem(name, value);
    } catch {
      // Storage is optional; keep the in-memory session usable.
    }
  },
  removeItem: (name) => {
    try {
      if (typeof window !== "undefined") window.localStorage.removeItem(name);
    } catch {
      // Storage is optional; the in-memory state is cleared independently.
    }
  },
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      expiresAt: null,
      remember: false,
      login: (user, remember) => {
        const ttl = remember ? SESSION_DURATION_MS * 7 : SESSION_DURATION_MS;
        const expiresAt = Date.now() + ttl;
        setCookie(AUTH_COOKIE_NAME, user.role, Math.ceil(ttl / 1000));
        set({ user, expiresAt, remember });
      },
      updateUser: (patch) => {
        set((state) => (state.user ? { user: { ...state.user, ...patch } } : state));
      },
      logout: () => {
        clearCookie(AUTH_COOKIE_NAME);
        set({ user: null, expiresAt: null, remember: false });
      },
      checkExpiry: () => {
        const { expiresAt, user } = get();
        if (user && expiresAt && Date.now() > expiresAt) {
          clearCookie(AUTH_COOKIE_NAME);
          set({ user: null, expiresAt: null, remember: false });
        }
      },
    }),
    {
      name: AUTH_STORAGE_KEY,
      storage: createJSONStorage(() => safeStorage),
      partialize: (s) => ({ user: s.user, expiresAt: s.expiresAt, remember: s.remember }),
    }
  )
);
