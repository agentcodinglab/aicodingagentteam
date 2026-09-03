import React from "react";
import { Box, Text } from "ink";

const commands = [
  { cmd: "/run <requirement>", desc: "Run full pipeline" },
  { cmd: "/quick <edit>", desc: "Quick edit" },
  { cmd: "/plan", desc: "Show DAG plan" },
  { cmd: "/continue", desc: "Resume parked task" },
  { cmd: "/verify", desc: "Run quality gate" },
  { cmd: "/backend <name>", desc: "Switch host CLI" },
  { cmd: "/report", desc: "Show quality report" },
  { cmd: "/exit", desc: "Exit" },
];

export function HelpPanel() {
  return (
    <Box flexDirection="column" borderStyle="round" borderColor="gray" paddingX={1} marginTop={0}>
      <Text bold color="gray">Commands</Text>
      {commands.map((c) => (
        <Box key={c.cmd}>
          <Text color="cyan">{c.cmd.padEnd(24)}</Text>
          <Text color="gray">{c.desc}</Text>
        </Box>
      ))}
    </Box>
  );
}