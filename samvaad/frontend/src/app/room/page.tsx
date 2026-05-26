'use client';

import React, { useEffect, useState, useRef } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, ScrollView, Platform } from 'react-native';
import { useSearchParams, useRouter } from 'next/navigation';

export default function RoomScreen() {
  const searchParams = useSearchParams();
  const router = useRouter();

  const room = searchParams.get('room') || 'Global';
  const identity = searchParams.get('identity') || 'Guest';
  const lowResource = searchParams.get('lowResource') === 'true';
  const e2ee = searchParams.get('e2ee') === 'true';

  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [audioMuted, setAudioMuted] = useState(false);
  const [videoDisabled, setVideoDisabled] = useState(false);
  const [isScreenSharing, setIsScreenSharing] = useState(false);

  const localVideoRef = useRef<HTMLVideoElement>(null);

  // Initialise local camera/microphone stream
  useEffect(() => {
    if (typeof window === 'undefined') return;

    async function initCamera() {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          video: { width: 640, height: 480, frameRate: 24 },
          audio: true,
        });
        setLocalStream(stream);
        if (localVideoRef.current) {
          localVideoRef.current.srcObject = stream;
        }
      } catch (err) {
        console.error('Failed to access media devices:', err);
      }
    }

    if (!videoDisabled) {
      initCamera();
    }

    return () => {
      if (localStream) {
        localStream.getTracks().forEach((track) => track.stop());
      }
    };
  }, []);

  // Handle Mute Audio toggle
  const toggleAudio = () => {
    if (localStream) {
      const audioTrack = localStream.getAudioTracks()[0];
      if (audioTrack) {
        audioTrack.enabled = audioMuted;
        setAudioMuted(!audioMuted);
      }
    }
  };

  // Handle Disable Video toggle
  const toggleVideo = () => {
    if (localStream) {
      const videoTrack = localStream.getVideoTracks()[0];
      if (videoTrack) {
        videoTrack.enabled = videoDisabled;
        setVideoDisabled(!videoDisabled);
      }
    }
  };

  const handleLeave = () => {
    if (localStream) {
      localStream.getTracks().forEach((track) => track.stop());
    }
    router.push('/');
  };

  return (
    <View style={styles.container}>
      {/* Header Info */}
      <View style={styles.header}>
        <View style={styles.headerLeft}>
          <Text style={styles.roomName}>{room.toUpperCase()}</Text>
          <View style={styles.badgeRow}>
            {e2ee && (
              <View style={[styles.badge, styles.badgeSecure]}>
                <Text style={styles.badgeText}>🛡️ E2EE MILITARY SECURE</Text>
              </View>
            )}
            {lowResource && (
              <View style={[styles.badge, styles.badgeLow]}>
                <Text style={styles.badgeText}>📉 LOW RESOURCE ACTIVE</Text>
              </View>
            )}
          </View>
        </View>
        <Text style={styles.identityText}>Joining as: {identity}</Text>
      </View>

      {/* Grid Video Container */}
      <ScrollView contentContainerStyle={styles.gridContainer}>
        {/* Local Participant Block */}
        <View style={styles.videoCard}>
          {typeof window !== 'undefined' && !videoDisabled ? (
            <video
              ref={localVideoRef}
              autoPlay
              playsInline
              muted
              style={{
                width: '100%',
                height: '100%',
                objectFit: 'cover',
                borderRadius: 12,
              }}
            />
          ) : (
            <View style={styles.videoPlaceholder}>
              <Text style={styles.placeholderLabel}>{identity.substring(0, 2).toUpperCase()}</Text>
            </View>
          )}
          <View style={styles.videoOverlay}>
            <Text style={styles.videoName}>{identity} (You)</Text>
          </View>
        </View>

        {/* Mock Participant 1 */}
        <View style={styles.videoCard}>
          <View style={[styles.videoPlaceholder, { backgroundColor: '#1d2430' }]}>
            <Text style={styles.placeholderLabel}>AR</Text>
          </View>
          <View style={styles.videoOverlay}>
            <Text style={styles.videoName}>Aarav Sharma</Text>
          </View>
        </View>

        {/* Mock Participant 2 */}
        <View style={styles.videoCard}>
          <View style={[styles.videoPlaceholder, { backgroundColor: '#201a2d' }]}>
            <Text style={styles.placeholderLabel}>VK</Text>
          </View>
          <View style={styles.videoOverlay}>
            <Text style={styles.videoName}>Vikram Kohli</Text>
          </View>
        </View>
      </ScrollView>

      {/* Controls Bar at bottom */}
      <View style={styles.controlsBar}>
        <TouchableOpacity
          style={[styles.controlBtn, audioMuted && styles.controlBtnAlert]}
          onPress={toggleAudio}
        >
          <Text style={styles.btnText}>{audioMuted ? '🎙️ Unmute' : '🎙️ Mute'}</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.controlBtn, videoDisabled && styles.controlBtnAlert]}
          onPress={toggleVideo}
        >
          <Text style={styles.btnText}>{videoDisabled ? '📹 Start Video' : '📹 Stop Video'}</Text>
        </TouchableOpacity>

        <TouchableOpacity style={[styles.controlBtn, styles.controlBtnLeave]} onPress={handleLeave}>
          <Text style={styles.btnTextLeave}>🔴 Leave Space</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0b0c0e',
    justifyContent: 'space-between',
  },
  header: {
    padding: 16,
    backgroundColor: '#121418',
    borderBottomWidth: 1,
    borderColor: '#1e222b',
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  headerLeft: {
    flexDirection: 'column',
    gap: 4,
  },
  roomName: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#ffffff',
    letterSpacing: 0.5,
  },
  badgeRow: {
    flexDirection: 'row',
    gap: 8,
  },
  badge: {
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
  },
  badgeSecure: {
    backgroundColor: '#102a1e',
    borderWidth: 1,
    borderColor: '#10b981',
  },
  badgeLow: {
    backgroundColor: '#2e1c18',
    borderWidth: 1,
    borderColor: '#f59e0b',
  },
  badgeText: {
    color: '#ffffff',
    fontSize: 9,
    fontWeight: 'bold',
  },
  identityText: {
    fontSize: 14,
    color: '#9ca3af',
  },
  gridContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'center',
    alignItems: 'center',
    gap: 16,
    padding: 24,
  },
  videoCard: {
    width: '100%',
    maxWidth: 360,
    height: 240,
    backgroundColor: '#121418',
    borderRadius: 12,
    borderWidth: 1,
    borderColor: '#1e222b',
    overflow: 'hidden',
    position: 'relative',
  },
  videoPlaceholder: {
    flex: 1,
    backgroundColor: '#16191f',
    justifyContent: 'center',
    alignItems: 'center',
  },
  placeholderLabel: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#3b82f6',
  },
  videoOverlay: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    padding: 8,
    backgroundColor: 'rgba(11, 12, 14, 0.75)',
  },
  videoName: {
    color: '#f3f4f6',
    fontSize: 12,
    fontWeight: '600',
  },
  controlsBar: {
    padding: 20,
    backgroundColor: '#121418',
    borderTopWidth: 1,
    borderColor: '#1e222b',
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    gap: 16,
  },
  controlBtn: {
    backgroundColor: '#22252a',
    borderRadius: 8,
    paddingHorizontal: 20,
    paddingVertical: 12,
    minWidth: 120,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#2d333f',
  },
  controlBtnAlert: {
    backgroundColor: '#3b1c1c',
    borderColor: '#ef4444',
  },
  controlBtnLeave: {
    backgroundColor: '#dc2626',
    borderColor: '#ef4444',
  },
  btnText: {
    color: '#f3f4f6',
    fontSize: 13,
    fontWeight: '600',
  },
  btnTextLeave: {
    color: '#ffffff',
    fontSize: 13,
    fontWeight: '600',
  },
});
