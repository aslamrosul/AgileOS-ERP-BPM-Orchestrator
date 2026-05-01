'use client';

import React, { createContext, useContext, useEffect, useState } from 'react';
import { useSocket, NotificationMessage } from '@/hooks/useSocket';
import { toast } from 'sonner';

interface WebSocketContextType {
  isConnected: boolean;
  connectionStatus: 'disconnected' | 'connecting' | 'connected' | 'error';
  lastNotification: NotificationMessage | null;
  notifications: NotificationMessage[];
  connect: () => void;
  disconnect: () => void;
  clearNotifications: () => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

interface WebSocketProviderProps {
  children: React.ReactNode;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({ children }) => {
  const [notifications, setNotifications] = useState<NotificationMessage[]>([]);

  const handleNotification = (notification: NotificationMessage) => {
    setNotifications(prev => [notification, ...prev.slice(0, 49)]); // Keep last 50 notifications
  };

  const handleConnect = () => {
    toast.success('Connected', {
      description: 'Real-time notifications are now active',
    });
  };

  const handleDisconnect = () => {
    toast.info('Disconnected', {
      description: 'Real-time notifications are offline',
    });
  };

  const handleError = (error: Event) => {
    console.error('WebSocket error:', error);
    toast.error('Connection Error', {
      description: 'Failed to establish real-time connection',
    });
  };

  const {
    isConnected,
    connectionStatus,
    lastNotification,
    connect,
    disconnect,
  } = useSocket({
    autoConnect: true,
    onNotification: handleNotification,
    onConnect: handleConnect,
    onDisconnect: handleDisconnect,
    onError: handleError,
  });

  const clearNotifications = () => {
    setNotifications([]);
  };

  const contextValue: WebSocketContextType = {
    isConnected,
    connectionStatus,
    lastNotification,
    notifications,
    connect,
    disconnect,
    clearNotifications,
  };

  return (
    <WebSocketContext.Provider value={contextValue}>
      {children}
    </WebSocketContext.Provider>
  );
};

export default WebSocketProvider;