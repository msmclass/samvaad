'use client';

import React, { useState } from 'react';
import { View, Text, StyleSheet, TextInput, TouchableOpacity } from 'react-native';
import { useRouter } from 'next/navigation';

export default function HomeScreen() {
  const [roomName, setRoomName] = useState('');
  const [identity, setIdentity] = useState('');
  const [isLowResource, setIsLowResource] = useState(false);
  const [isE2EE, setIsE2EE] = useState(true);
  const router = useRouter();

  const handleJoin = () => {
    if (!roomName || !identity) return;
    const cleanRoom = encodeURIComponent(roomName.trim());
    const cleanIdentity = encodeURIComponent(identity.trim());
    router.push(`/room?room=${cleanRoom}&identity=${cleanIdentity}&lowResource=${isLowResource}&e2ee=${isE2EE}`);
  };

  return (
    <View style={styles.container}>
      <View style={styles.card}>
        <Text style={styles.title}>SAMVAAD MEET</Text>
        <Text style={styles.subtitle}>Sovereign virtual meeting space — 100% Free & Secure</Text>

        <View style={styles.inputContainer}>
          <Text style={styles.label}>Your Name</Text>
          <TextInput
            style={styles.input}
            value={identity}
            onChangeText={setIdentity}
            placeholder="Enter your name"
            placeholderTextColor="#4b5563"
          />
        </View>

        <View style={styles.inputContainer}>
          <Text style={styles.label}>Room Name</Text>
          <TextInput
            style={styles.input}
            value={roomName}
            onChangeText={setRoomName}
            placeholder="Enter room identifier"
            placeholderTextColor="#4b5563"
          />
        </View>

        <View style={styles.toggleRow}>
          <TouchableOpacity 
            style={[styles.toggleBtn, isE2EE && styles.toggleBtnActive]}
            onPress={() => setIsE2EE(!isE2EE)}
          >
            <Text style={[styles.toggleText, isE2EE && styles.toggleTextActive]}>
              🛡️ E2EE Security: {isE2EE ? 'ON' : 'OFF'}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity 
            style={[styles.toggleBtn, isLowResource && styles.toggleBtnActive]}
            onPress={() => setIsLowResource(!isLowResource)}
          >
            <Text style={[styles.toggleText, isLowResource && styles.toggleTextActive]}>
              📉 Low Resource: {isLowResource ? 'ON' : 'OFF'}
            </Text>
          </TouchableOpacity>
        </View>

        <TouchableOpacity 
          style={[styles.joinButton, (!roomName || !identity) && styles.joinButtonDisabled]}
          onPress={handleJoin}
          disabled={!roomName || !identity}
        >
          <Text style={styles.joinButtonText}>Enter Meeting Space</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0b0c0e',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 16,
  },
  card: {
    backgroundColor: '#121418',
    borderRadius: 16,
    padding: 32,
    width: '100%',
    maxWidth: 480,
    borderWidth: 1,
    borderColor: '#1e222b',
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#3b82f6',
    textAlign: 'center',
    marginBottom: 8,
    letterSpacing: 1.5,
  },
  subtitle: {
    fontSize: 14,
    color: '#9ca3af',
    textAlign: 'center',
    marginBottom: 32,
  },
  inputContainer: {
    marginBottom: 20,
  },
  label: {
    fontSize: 12,
    fontWeight: '600',
    color: '#9ca3af',
    textTransform: 'uppercase',
    marginBottom: 8,
    letterSpacing: 0.5,
  },
  input: {
    backgroundColor: '#1a1d24',
    borderWidth: 1,
    borderColor: '#262c38',
    borderRadius: 8,
    padding: 12,
    color: '#f3f4f6',
    fontSize: 16,
  },
  toggleRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 32,
    gap: 12,
  },
  toggleBtn: {
    flex: 1,
    backgroundColor: '#1a1d24',
    borderWidth: 1,
    borderColor: '#262c38',
    borderRadius: 8,
    padding: 10,
    alignItems: 'center',
  },
  toggleBtnActive: {
    borderColor: '#3b82f6',
    backgroundColor: '#142036',
  },
  toggleText: {
    fontSize: 11,
    fontWeight: '600',
    color: '#9ca3af',
    textAlign: 'center',
  },
  toggleTextActive: {
    color: '#60a5fa',
  },
  joinButton: {
    backgroundColor: '#2563eb',
    borderRadius: 8,
    padding: 16,
    alignItems: 'center',
  },
  joinButtonDisabled: {
    backgroundColor: '#1e293b',
    opacity: 0.5,
  },
  joinButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '600',
  },
});
