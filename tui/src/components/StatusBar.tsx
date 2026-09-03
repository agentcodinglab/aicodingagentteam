import React from "react";
import { Box, Text } from "ink";
import Spinner from "ink-spinner";
import { useStore } from "../store.js";

export function StatusBar() {
  const { connected, address, running, currentPhase, currentRole, score, passed } = useStore();

  return (
    <Box flexDirection="column" borderStyle="round" borderColor="cyan" paddingX={1}>
      <Box>
        <Text color="green">{connected ? "●" : "○"}</Text>
        <Text color="gray"> {address}</Text>
        {running && (
          <>
            <Text> </Text>
            <Spinner type="dots" />
            <Text color="yellow"> {currentPhase}/{currentRole}</Text>
          </>
        )}
      </Box>
      {score > 0 && (
        <Box marginTop={0}>
          <Text>
            Quality: <Text color={passed ? "green" : "red"}>{score}</Text>
            <Text color={passed ? "green" : "red"}> {passed ? "PASS" : "FAIL"}</Text>
          </Text>
        </Box>
      )}
    </Box>
  );
}