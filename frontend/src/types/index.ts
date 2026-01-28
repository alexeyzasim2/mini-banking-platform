export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  created_at: string;
}

export interface Account {
  id: string;
  user_id: string;
  currency: string;
  balance_cents: number;
  created_at: string;
}

export interface Transaction {
  id: string;
  type: 'transfer' | 'exchange';
  from_user_id: string;
  to_user_id?: string;
  amount_cents: number;
  currency: string;
  description: string;
  created_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
}

export interface TransferRequest {
  to_user_id: string;
  currency: string;
  amount_cents: number;
}

export interface ExchangeRequest {
  from_currency: string;
  amount_cents: number;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ErrorResponse {
  error: string;
}

export const centsToDollars = (cents: number): string => {
  return (cents / 100).toFixed(2);
};

export const dollarsToCents = (dollars: string | number): number => {
  const amount = typeof dollars === 'string' ? parseFloat(dollars) : dollars;
  return Math.round(amount * 100);
};
