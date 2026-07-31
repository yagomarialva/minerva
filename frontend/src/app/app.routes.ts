import { Routes } from '@angular/router';
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MinervaService } from './services/minerva.service';

@Component({
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="search-container">
      <h2>Buscar Acervo</h2>
      <div class="search-bar">
        <input 
          type="text" 
          [(ngModel)]="searchQuery" 
          (keyup.enter)="onSearch()" 
          placeholder="Busque por título ou autor..." 
          class="neon-input"
        />
        <button (click)="onSearch()" class="neon-btn" [disabled]="loading">
          {{ loading ? 'Buscando...' : 'Pesquisar' }}
        </button>
      </div>

      <div class="results-grid" *ngIf="results.length > 0">
        <div class="book-card" *ngFor="let book of results">
          <div class="cover">
             <span class="extension">{{ book.extension || 'EPUB' }}</span>
          </div>
          <div class="info">
            <h3>{{ book.title || 'Sem Título' }}</h3>
            <p>{{ book.author || 'Desconhecido' }}</p>
            <p class="meta">{{ book.filesize || 'Desconhecido' }} • {{ book.language || '?' }}</p>
            <button class="neon-btn-small" (click)="download(book)">Baixar</button>
          </div>
        </div>
      </div>
      
      <div class="no-results" *ngIf="!loading && searched && results.length === 0">
        <p>Nenhum livro encontrado.</p>
      </div>
    </div>
  `,
  styles: [`
    .search-container {
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
    }
    .search-bar {
      display: flex;
      gap: 1rem;
    }
    .neon-input {
      flex: 1;
      padding: 0.8rem 1rem;
      border-radius: 8px;
      border: 1px solid rgba(138, 43, 226, 0.3);
      background: rgba(0,0,0,0.4);
      color: #fff;
      font-size: 1rem;
      outline: none;
      transition: all 0.3s ease;
    }
    .neon-input:focus {
      border-color: rgba(138, 43, 226, 1);
      box-shadow: 0 0 10px rgba(138, 43, 226, 0.5);
    }
    .neon-btn {
      padding: 0.8rem 1.5rem;
      border-radius: 8px;
      border: 1px solid rgba(138, 43, 226, 0.8);
      background: rgba(138, 43, 226, 0.2);
      color: #e0b0ff;
      font-weight: 300;
      cursor: pointer;
      transition: all 0.3s ease;
    }
    .neon-btn:hover:not([disabled]) {
      background: rgba(138, 43, 226, 0.5);
      box-shadow: 0 0 15px rgba(138, 43, 226, 0.8);
    }
    .neon-btn:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
    .neon-btn-small {
      margin-top: 0.5rem;
      padding: 0.4rem 0.8rem;
      font-size: 0.8rem;
      border-radius: 4px;
      border: 1px solid rgba(138, 43, 226, 0.8);
      background: transparent;
      color: #e0b0ff;
      cursor: pointer;
      transition: all 0.3s ease;
    }
    .neon-btn-small:hover {
      background: rgba(138, 43, 226, 0.3);
    }
    .results-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
      gap: 1.5rem;
    }
    .book-card {
      background: rgba(255, 255, 255, 0.05);
      border-radius: 12px;
      overflow: hidden;
      border: 1px solid rgba(255,255,255,0.1);
      display: flex;
      flex-direction: column;
    }
    .cover {
      height: 120px;
      background: linear-gradient(135deg, rgba(138,43,226,0.2) 0%, rgba(0,0,0,0.6) 100%);
      display: flex;
      align-items: center;
      justify-content: center;
      border-bottom: 1px solid rgba(255,255,255,0.1);
    }
    .extension {
      font-size: 1.2rem;
      font-weight: bold;
      color: rgba(255,255,255,0.5);
      letter-spacing: 2px;
    }
    .info {
      padding: 1rem;
      flex: 1;
      display: flex;
      flex-direction: column;
    }
    .info h3 {
      font-size: 1rem;
      margin: 0 0 0.5rem 0;
      color: #fff;
    }
    .info p {
      font-size: 0.8rem;
      color: rgba(255,255,255,0.7);
      margin: 0 0 0.3rem 0;
    }
    .info .meta {
      font-size: 0.7rem;
      color: rgba(138, 43, 226, 0.8);
      margin-bottom: auto;
    }
    .no-results {
      text-align: center;
      padding: 3rem;
      color: rgba(255,255,255,0.5);
    }
  `]
})
export class SearchComponent {
  searchQuery = '';
  results: any[] = [];
  loading = false;
  searched = false;

  constructor(private minerva: MinervaService) {}

  onSearch() {
    if (!this.searchQuery.trim()) return;
    
    this.loading = true;
    this.searched = true;
    this.results = [];
    
    this.minerva.searchBooks(this.searchQuery).subscribe({
      next: (data) => {
        this.results = data || [];
        this.loading = false;
      },
      error: (err) => {
        console.error('Search failed', err);
        this.loading = false;
      }
    });
  }

  download(book: any) {
    this.minerva.queueDownload(book).subscribe({
      next: () => {
        alert(`"${book.title}" enviado para a fila de downloads!`);
      },
      error: (err) => console.error('Erro ao baixar:', err)
    });
  }
}

import { OnInit, ChangeDetectorRef } from '@angular/core';

@Component({ 
  standalone: true, 
  imports: [CommonModule],
  template: `
    <h2>Minha Biblioteca</h2>
    <div class="results-grid" *ngIf="books.length > 0">
      <div class="book-card" *ngFor="let book of books">
        <div class="cover">
           <span class="extension">{{ book.extension }}</span>
        </div>
        <div class="info">
          <h3>{{ book.title }}</h3>
          <p>{{ book.author }}</p>
          <p class="meta">{{ book.filesize }} • Baixado</p>
          <div class="actions">
            <button class="neon-btn-small" (click)="downloadFile(book)">Baixar para o PC</button>
            <button class="neon-btn-small convert-btn" *ngIf="book.extension !== 'EPUB' && !converting[book.id]" (click)="convertToEpub(book)">Converter para EPUB</button>
            <span class="meta" *ngIf="converting[book.id]">Convertendo...</span>
          </div>
        </div>
      </div>
    </div>
    <div class="no-results" *ngIf="books.length === 0">
      <p>Sua biblioteca está vazia.</p>
    </div>
  ` 
})
export class LibraryComponent implements OnInit {
  books: any[] = [];
  converting: { [id: number]: boolean } = {};

  constructor(private minerva: MinervaService, private cdr: ChangeDetectorRef) {}

  ngOnInit() {
    this.loadLibrary();
    this.minerva.getMessages().subscribe(msg => {
      if (msg.type === 'UPDATE_QUEUE') {
        this.loadLibrary();
      }
    });
  }

  loadLibrary() {
    this.minerva.getLibrary().subscribe(data => {
      this.books = data || [];
      // Reset converting state for loaded books
      this.books.forEach(b => { if(b.extension === 'EPUB') delete this.converting[b.id]; });
      this.cdr.detectChanges();
    });
  }
  
  downloadFile(book: any) {
    window.location.href = `/api/download-file?id=${book.id}`;
  }

  convertToEpub(book: any) {
    this.converting[book.id] = true;
    this.minerva.convertBook(book.id).subscribe();
  }
}

@Component({ 
  standalone: true, 
  imports: [CommonModule],
  template: `
    <h2>Fila de Downloads</h2>
    <div class="queue-list" *ngIf="queue.length > 0">
      <div class="queue-item" *ngFor="let item of queue">
        <div class="q-info">
          <strong>{{ item.title }}</strong>
          <span>Status: {{ item.status }}</span>
        </div>
        <div class="progress-bar">
          <div class="progress" [style.width.%]="progressMap[item.id] || 0"></div>
        </div>
      </div>
    </div>
    <div class="no-results" *ngIf="queue.length === 0">
      <p>Nenhum download em andamento.</p>
    </div>
  `,
  styles: [`
    .queue-list { display: flex; flex-direction: column; gap: 1rem; }
    .queue-item { background: rgba(0,0,0,0.3); padding: 1rem; border-radius: 8px; border: 1px solid rgba(138,43,226,0.3); }
    .q-info { display: flex; justify-content: space-between; margin-bottom: 0.5rem; color: #fff; }
    .progress-bar { height: 8px; background: rgba(255,255,255,0.1); border-radius: 4px; overflow: hidden; }
    .progress { height: 100%; background: #8a2be2; transition: width 0.3s ease; }
  `]
})
export class QueueComponent implements OnInit {
  queue: any[] = [];
  progressMap: { [id: number]: number } = {};

  constructor(private minerva: MinervaService, private cdr: ChangeDetectorRef) {}

  ngOnInit() {
    this.loadQueue();
    this.minerva.getMessages().subscribe(msg => {
      if (msg.type === 'UPDATE_QUEUE') {
        this.loadQueue();
      } else if (msg.type === 'PROGRESS') {
        this.progressMap[msg.payload.id] = msg.payload.percent;
        this.cdr.detectChanges();
      }
    });
  }

  loadQueue() {
    this.minerva.getQueue().subscribe(data => {
      this.queue = data || [];
      this.cdr.detectChanges();
    });
  }
}

@Component({ 
  standalone: true, 
  imports: [CommonModule, FormsModule],
  template: `
    <h2>Configurações do Kindle</h2>
    <div class="settings-form">
      <label>Email do seu Kindle (ex: seunome&#64;kindle.com)</label>
      <input type="email" [(ngModel)]="settings.kindleEmail" class="neon-input" />
      
      <label>Seu Email SMTP (Remetente)</label>
      <input type="email" [(ngModel)]="settings.smtpEmail" class="neon-input" />

      <label>Senha do SMTP</label>
      <input type="password" [(ngModel)]="settings.smtpPassword" class="neon-input" />

      <button class="neon-btn" (click)="save()">Salvar</button>
    </div>
  `,
  styles: [`
    .settings-form { display: flex; flex-direction: column; gap: 1rem; max-width: 400px; }
    label { color: #e0b0ff; font-size: 0.9rem; }
    .neon-input { padding: 0.8rem; border-radius: 8px; border: 1px solid rgba(138,43,226,0.3); background: rgba(0,0,0,0.4); color: #fff; }
    .neon-btn { margin-top: 1rem; padding: 0.8rem; border-radius: 8px; border: 1px solid #8a2be2; background: rgba(138,43,226,0.2); color: #fff; cursor: pointer; }
  `]
})
export class SettingsComponent implements OnInit {
  settings: any = { kindleEmail: '', smtpEmail: '', smtpPassword: '' };
  
  constructor(private minerva: MinervaService) {}

  ngOnInit() {
    this.minerva.getSettings().subscribe(data => {
      if(data) this.settings = data;
    });
  }

  save() {
    this.minerva.saveSettings(this.settings).subscribe(() => {
      alert('Configurações salvas!');
    });
  }
}

export const routes: Routes = [
  { path: 'search', component: SearchComponent },
  { path: 'library', component: LibraryComponent },
  { path: 'queue', component: QueueComponent },
  { path: 'settings', component: SettingsComponent },
  { path: '', redirectTo: '/library', pathMatch: 'full' }
];
