import { api } from "../../lib/apiClient";

type UserWire = {
  id: string;
  email: string;
  locale: string;
  dir_pref: string;
};

export type User = {
  id: string;
  email: string;
  locale: string;
  dirPref: string;
};

const toUser = (wire: UserWire): User => ({
  id: wire.id,
  email: wire.email,
  locale: wire.locale,
  dirPref: wire.dir_pref,
});

export const fetchMe = async (): Promise<User> => toUser(await api.get<UserWire>("/me"));

export const login = async (email: string, password: string): Promise<User> =>
  toUser(await api.post<UserWire>("/auth/login", { email, password }));

export const register = async (email: string, password: string, invite: string): Promise<User> =>
  toUser(await api.post<UserWire>("/auth/register", { email, password, invite }));

export const logout = (): Promise<void> => api.post<void>("/auth/logout");
