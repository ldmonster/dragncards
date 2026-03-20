import React, { useEffect, useRef, ReactNode } from "react";
import { Socket } from "phoenix";

import SocketContext from "../contexts/SocketContext";

const SocketProvider = ({
  wsUrl,
  options,
  children,
}: {
  wsUrl: string;
  options: object | (() => object);
  children: ReactNode;
}) => {
  // Keep a mutable ref so the params function always returns the latest options
  // without recreating the Socket or reconnecting the WebSocket.
  const optionsRef = useRef(options);
  optionsRef.current = options;

  // Create the Socket exactly once. Using a function ref means reconnections
  // (e.g. after a network drop) will automatically pick up the latest params.
  const socketRef = useRef<Socket | null>(null);
  if (socketRef.current === null) {
    socketRef.current = new Socket(wsUrl, { params: () => optionsRef.current });
  }

  useEffect(() => {
    socketRef.current!.connect();
    return () => {
      socketRef.current!.disconnect();
    };
  }, []); // connect once on mount, disconnect on unmount

  return (
    <SocketContext.Provider value={socketRef.current}>{children}</SocketContext.Provider>
  );
};

export default SocketProvider;
