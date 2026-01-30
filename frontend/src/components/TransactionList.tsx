import { useState } from 'react';
import type { Transaction } from '../types';
import { TransactionDetailModal } from './TransactionDetailModal';
import { centsToDollars } from '../types';

interface Props {
  transactions: Transaction[] | null | undefined;
}

export const TransactionList = ({ transactions }: Props) => {
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null);
  const [showModal, setShowModal] = useState(false);

  const handleViewDetails = (transaction: Transaction) => {
    setSelectedTransaction(transaction);
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setSelectedTransaction(null);
  };
  const formatDate = (date: string) => {
    return new Date(date).toLocaleString();
  };

  const getTypeColor = (type: string) => {
    if (type === 'transfer') {
      return 'bg-blue-100 text-blue-800';
    }
    if (type === 'exchange') {
      return 'bg-green-100 text-green-800';
    }
    return 'bg-gray-100 text-gray-800';
  };

  const list = transactions || [];

  if (list.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        No transactions yet
      </div>
    );
  }

  return (
    <>
      <TransactionDetailModal
        isOpen={showModal}
        transaction={selectedTransaction}
        onClose={handleCloseModal}
      />

      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Type
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Amount
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Description
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Date
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {list.map((transaction) => (
              <tr
                key={transaction.id}
                className="hover:bg-gray-50 cursor-pointer transition-colors"
                onClick={() => handleViewDetails(transaction)}
              >
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2 py-1 text-xs font-semibold rounded-full ${getTypeColor(transaction.type)}`}>
                    {transaction.type}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {centsToDollars(transaction.amount_cents)} {transaction.currency}
                </td>
                <td className="px-6 py-4 text-sm text-gray-500">
                  {transaction.description}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {formatDate(transaction.created_at)}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleViewDetails(transaction);
                    }}
                    className="text-blue-600 hover:text-blue-800 font-medium"
                  >
                    View Receipt
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
};

