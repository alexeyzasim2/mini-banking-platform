import { useState, useEffect } from 'react';
import { accountsApi, transactionsApi } from '../services/api';
import { TransferForm } from '../components/TransferForm';
import { ExchangeForm } from '../components/ExchangeForm';
import { TransactionList } from '../components/TransactionList';
import type { Account, Transaction } from '../types';
import { centsToDollars } from '../types';

export const Dashboard = () => {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    try {
      const [accountsRes, transactionsRes] = await Promise.all([
        accountsApi.getAccounts(),
        transactionsApi.getTransactions({ limit: 5 }),
      ]);
      setAccounts(accountsRes.data.accounts);
      setTransactions(transactionsRes.data.transactions);
    } catch (error) {
      console.error('Failed to load data', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
      </div>

      {/* Account Balances */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {accounts.map((account) => (
          <div key={account.id} className="bg-white rounded-lg shadow p-6">
            <div className="flex justify-between items-center">
              <div>
                <p className="text-sm text-gray-500">{account.currency} Account</p>
                <p className="text-3xl font-bold text-gray-900">
                  {account.currency === 'USD' ? '$' : '€'}
                  {centsToDollars(account.balance_cents)}
                </p>
              </div>
              <div className={`text-4xl ${account.currency === 'USD' ? 'text-green-500' : 'text-blue-500'}`}>
                {account.currency === 'USD' ? '$' : '€'}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Forms */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-stretch">
        <div className="bg-white rounded-lg shadow p-6 flex flex-col h-full">
          <h2 className="text-xl font-semibold mb-4">Transfer Money</h2>
          <TransferForm accounts={accounts} onSuccess={loadData} />
        </div>
        <div className="bg-white rounded-lg shadow p-6 flex flex-col h-full">
          <h2 className="text-xl font-semibold mb-4">Exchange Currency</h2>
          <ExchangeForm onSuccess={loadData} />
        </div>
      </div>

      {/* Recent Transactions */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4">Recent Transactions (Last 5)</h2>
        <TransactionList transactions={transactions} />
      </div>
    </div>
  );
};

