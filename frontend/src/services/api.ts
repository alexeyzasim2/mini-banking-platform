import axios from 'axios';
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  User,
  Account,
  Transaction,
  TransferRequest,
  ExchangeRequest,
} from '../types';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const authApi = {
  login: (data: LoginRequest) =>
    api.post<AuthResponse>('/auth/login', data),
  
  register: (data: RegisterRequest) =>
    api.post<AuthResponse>('/auth/register', data),
  
  getMe: () =>
    api.get<User>('/auth/me'),
};

export const accountsApi = {
  getAccounts: () =>
    api.get<{ accounts: Account[] }>('/accounts'),
  
  getBalance: (accountId: string) =>
    api.get<{ balance: string }>(`/accounts/${accountId}/balance`),
};

export const transactionsApi = {
  transfer: (data: TransferRequest) =>
    api.post<Transaction>('/transactions/transfer', data),
  
  exchange: (data: ExchangeRequest) =>
    api.post<Transaction>('/transactions/exchange', data),
  
  getTransactions: (params?: { type?: string; page?: number; limit?: number }) =>
    api.get<{ transactions: Transaction[]; page: number; limit: number; total: number }>('/transactions', { params }),
};

export default api;

