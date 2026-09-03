import React from "react";
import { Box, Text } from "ink";
import { useStore } from "../store.js";

export function AuditLog() {
  const { auditLog } = useStore();

  if (auditLog.length === 0) {
    return null;
  }

  const recent = auditLog.slice(-8);

  return (
    <Box flexDirection="column" borderStyle="round" borderColor="gray" paddingX={1} marginTop={0}>
      <Text bold color="gray">Audit</Text>
      {recent.map((entry, i) => (
        <Box key={i}>
          <Text color="gray">{entry.ts.slice(11, 19)} </Text>
          <Text color="cyan">{entry.type.padEnd(16)} </Text>
          <Text color="yellow">{entry.agent.padEnd(12)} </Text>
          <Text color={entry.result === "pass" || entry.result === "accept" ? "green" : "red"}>
            {entry.result}
          </Text>
        </Box>
      ))}
    </Box>
  );
}