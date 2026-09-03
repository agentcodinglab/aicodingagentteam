import React from "react";
import { Box, Text } from "ink";
import { useStore } from "../store.js";

export function ResultPanel() {
  const { score, passed, blocking, advisory, artifacts } = useStore();

  if (score === 0 && artifacts.length === 0) {
    return null;
  }

  return (
    <Box flexDirection="column" borderStyle="round" borderColor={passed ? "green" : "red"} paddingX={1} marginTop={0}>
      <Text bold>
        Result: <Text color={passed ? "green" : "red"}>{passed ? "PASS" : "FAIL"}</Text>
        <Text> (score: {score})</Text>
      </Text>
      {blocking.length > 0 && (
        <Box marginTop={0}>
          <Text color="red">Blocking: {blocking.join(", ")}</Text>
        </Box>
      )}
      {advisory.length > 0 && (
        <Box marginTop={0}>
          <Text color="yellow">Advisory: {advisory.join(", ")}</Text>
        </Box>
      )}
      {artifacts.length > 0 && (
        <Box marginTop={0} flexDirection="column">
          <Text color="gray">Artifacts:</Text>
          {artifacts.slice(0, 5).map((a, i) => (
            <Text key={i} color="gray">  {a}</Text>
          ))}
          {artifacts.length > 5 && <Text color="gray">  ... +{artifacts.length - 5} more</Text>}
        </Box>
      )}
    </Box>
  );
}