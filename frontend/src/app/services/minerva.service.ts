import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Subject, Observable } from 'rxjs';

export interface WsMessage {
  type: string;
  payload: any;
}

@Injectable({
  providedIn: 'root'
})
export class MinervaService {
  private socket!: WebSocket;
  private messagesSubject = new Subject<WsMessage>();

  constructor(private http: HttpClient) {
    this.connectWs();
  }

  private connectWs() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    this.socket = new WebSocket(`${protocol}//${host}/ws/`);
    
    this.socket.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data);
        this.messagesSubject.next(msg);
      } catch (e) {
        console.error('Error parsing WS message', e);
      }
    };
    
    this.socket.onclose = () => {
      console.log('WS connection closed. Reconnecting in 5s...');
      setTimeout(() => this.connectWs(), 5000);
    };
  }

  getMessages(): Observable<WsMessage> {
    return this.messagesSubject.asObservable();
  }

  searchBooks(query: string): Observable<any[]> {
    return this.http.get<any[]>(`/api/search?q=${encodeURIComponent(query)}`);
  }

  getLibrary(): Observable<any[]> {
    return this.http.get<any[]>('/api/library');
  }

  getQueue(): Observable<any[]> {
    return this.http.get<any[]>('/api/queue');
  }

  queueDownload(book: any): Observable<any> {
    return this.http.post<any>('/api/download', book);
  }

  convertBook(id: number): Observable<any> {
    return this.http.post<any>(`/api/convert?id=${id}`, {});
  }

  getSettings(): Observable<any> {
    return this.http.get<any>('/api/settings');
  }

  saveSettings(settings: any): Observable<any> {
    return this.http.post<any>('/api/settings', settings);
  }
}
