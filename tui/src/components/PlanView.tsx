import React from "react";
import { Box, Text } from "ink";
import { useStore } from "../store.js";

const STATUS_COLORS: Record<string, string> = {
  pending: "gray",
  running: "yellow",
  completed: "green",
  failed: "red",
  parked: "magenta",
};

const STATUS_ICONS: Record<string, string> = {
  pending: "○",
  running: "◐",
  completed: "●",
  failed: "✗",
  parked: "⏸",
};

export function PlanView() {
  const { nodes } = useStore();

  if (nodes.length === 0) {
    return null;
  }

  return (
    <Box flexDirection="column" borderStyle="round" borderColor="gray" paddingX={1} marginTop={0}>
      <Text bold>Plan DAG</Text>
      {nodes.map((node) => (
        <Box key={node.id}>
          <Text color={STATUS_COLORS[node.status]}>{STATUS_ICONS[node.status]} </Text>
          <Text color={STATUS_COLORS[node.status]}>
            {node.id.padEnd(16)}
          </Text>
          <Text color="gray"> {node.role.padEnd(12)}</Text>
          <Text color="gray">{node.phase}</Text>
        </Box>
      ))}
    </Box>
  );
}