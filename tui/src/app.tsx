import React, { useEffect, useState } from "react";
import { Box, Text, useInput, useApp } from "ink";
import { StatusBar } from "./components/StatusBar.js";
import { PlanView } from "./components/PlanView.js";
import { AuditLog } from "./components/AuditLog.js";
import { ResultPanel } from "./components/ResultPanel.js";
import { HelpPanel } from "./components/HelpPanel.js";
import { useStore } from "./store.js";
import { useCommands } from "./hooks/useCommands.js";
import { CoordinatorClient } from "./grpc/client.js";
import { MockClient } from "./grpc/mock.js";
import type { ClientInterface } from "./grpc/client.js";

interface AppProps {
  demo?: boolean;
}

function App({ demo = false }: AppProps) {
  const [input, setInput] = useState("");
  const [history, setHistory] = useState<string[]>([]);
  const [client, setClient] = useState<ClientInterface | null>(null);
  const store = useStore();
  const { exit } = useApp();

  useEffect(() => {
    let c: ClientInterface;
    if (demo) {
      c = new MockClient();
      store.setConnected(true);
      store.setAddress("demo://mock");
    } else {
      const host = process.env.AICODINGAGENTTEAM_HOST || "localhost";
      const port = parseInt(process.env.AICODINGAGENTTEAM_PORT || "8080", 10);
      c = new CoordinatorClient(host, port);
      store.setConnected(true);
      store.setAddress(`${host}:${port}`);
    }
    setClient(c);
  }, [demo, store]);

  const { execute } = useCommands(client);

  useInput((ch, key) => {
    if (key.ctrl && ch === "c") {
      exit();
      return;
    }
    if (key.return) {
      if (input.trim()) {
        setHistory((h) => [...h, input]);
        if (client) {
          execute(input);
        }
        setInput("");
      }
      return;
    }
    if (key.backspace || key.delete) {
      setInput((i) => i.slice(0, -1));
      return;
    }
    setInput((i) => i + ch);
  });

  const lastEntry = history.length > 0 ? history[history.length - 1] : "";

  return (
    <Box flexDirection="column" padding={1}>
      <StatusBar />
      <Box marginTop={0} flexDirection="row">
        <Box flexDirection="column" flexGrow={1} marginRight={1}>
          <PlanView />
          <ResultPanel />
          <AuditLog />
        </Box>
        <Box flexDirection="column" width={32}>
          <HelpPanel />
        </Box>
      </Box>
      <Box marginTop={1} borderStyle="round" borderColor="gray" paddingX={1}>
        <Text color="green">{">"} </Text>
        <Text>{input}</Text>
        <Text color="gray">_</Text>
      </Box>
      {lastEntry && (
        <Box marginTop={0}>
          <Text color="gray">last: {lastEntry}</Text>
        </Box>
      )}
      {store.message && (
        <Box marginTop={0}>
          <Text color="yellow">{store.message}</Text>
        </Box>
      )}
    </Box>
  );
}

export default App;